package ofx

import (
	"fmt"
	"strings"

	"github.com/bracomil/ofxgo/internal/sgml"
	"github.com/bracomil/ofxgo/internal/textcodec"
)

type v102Parser struct{}

func newV102Parser() VersionParser    { return &v102Parser{} }
func (p *v102Parser) Version() string { return "102" }

type parseContext struct {
	warnings []Warning
}

func (ctx *parseContext) Warn(w Warning) {
	ctx.warnings = append(ctx.warnings, w)
}

func (p *v102Parser) Parse(header Header, body []byte) (*Document, error) {
	ctx := &parseContext{}
	decoded := textcodec.Decode(body, header.Encoding, header.Charset)
	tree, err := sgml.Parse(decoded)
	if err != nil {
		return nil, fmt.Errorf("ofx 1.0.2: parse SGML: %w", err)
	}

	ofxRoot := tree.Child("OFX")
	if ofxRoot == nil {
		return nil, fmt.Errorf("ofx 1.0.2: OFX root missing")
	}

	doc := &Document{Header: header, RawVersion: "1.0.2"}
	stmtNodes := ofxRoot.Descendants("STMTRS")
	for _, n := range stmtNodes {
		st, err := parseStatement102(ctx, n)
		if err != nil {
			return nil, err
		}
		doc.Statements = append(doc.Statements, st)
	}
	if len(doc.Statements) == 0 {
		return nil, fmt.Errorf("ofx 1.0.2: no bank statements (STMTRS) found")
	}
	doc.Warnings = ctx.warnings
	return doc, nil
}

func parseStatement102(ctx *parseContext, n *sgml.Node) (BankStatement, error) {
	st := BankStatement{Currency: n.Value("CURDEF"), Raw: nodeValues(n)}
	if acct := n.Child("BANKACCTFROM"); acct != nil {
		st.Account = BankAccount{
			BankID: acct.Value("BANKID"), BranchID: acct.Value("BRANCHID"),
			AccountID: acct.Value("ACCTID"), AccountType: acct.Value("ACCTTYPE"),
			AccountKey: acct.Value("ACCTKEY"),
		}
	}

	if list := n.Child("BANKTRANLIST"); list != nil {
		st.StartDate = parseOptionalOFXDate(list.Value("DTSTART"))
		if st.StartDate.Present && !st.StartDate.Valid {
			ctx.Warn(Warning{
				Field:   "BANKTRANLIST.DTSTART",
				Value:   list.Value("DTSTART"),
				Message: "INVALID_OPTIONAL_DATE",
			})
		}

		st.EndDate = parseOptionalOFXDate(list.Value("DTEND"))
		if st.EndDate.Present && !st.EndDate.Valid {
			ctx.Warn(Warning{
				Field:   "BANKTRANLIST.DTEND",
				Value:   list.Value("DTEND"),
				Message: "INVALID_OPTIONAL_DATE",
			})
		}

		for _, txn := range list.ChildrenNamed("STMTTRN") {
			t, err := parseTransaction102(ctx, txn)
			if err != nil {
				return st, err
			}
			st.Transactions = append(st.Transactions, t)
		}
	}

	if b := n.Child("LEDGERBAL"); b != nil {
		rawDate := b.Value("DTASOF")

		bal := &Balance{
			Amount: strings.TrimSpace(b.Value("BALAMT")),
			AsOf:   parseOptionalOFXDate(rawDate),
		}

		if bal.AsOf.Present && !bal.AsOf.Valid {
			ctx.Warn(Warning{
				Message: "INVALID_OPTIONAL_DATE",
				Field:   "LEDGERBAL.DTASOF",
				Value:   rawDate,
			})
		}

		st.LedgerBalance = bal
	}
	return st, nil
}

func parseTransaction102(ctx *parseContext, n *sgml.Node) (Transaction, error) {
	t := Transaction{
		Type: n.Value("TRNTYPE"), Amount: strings.TrimSpace(n.Value("TRNAMT")), FITID: n.Value("FITID"),
		CorrectFITID: n.Value("CORRECTFITID"), CorrectAction: n.Value("CORRECTACTION"), ServerID: n.Value("SRVRTID"),
		CheckNumber: n.Value("CHECKNUM"), Reference: n.Value("REFNUM"), SIC: n.Value("SIC"), PayeeID: n.Value("PAYEEID"),
		Name: n.Value("NAME"), ExtendedName: n.Value("EXTDNAME"), Memo: n.Value("MEMO"), Raw: nodeValues(n),
	}
	var err error
	if t.PostedAt, err = parseOFXDate(n.Value("DTPOSTED")); err != nil {
		return t, fmt.Errorf("ofx: transaction FITID %q DTPOSTED: %w", t.FITID, err)
	}

	t.UserAt = parseOptionalOFXDate(n.Value("DTUSER"))
	if t.UserAt.Present && !t.UserAt.Valid {
		ctx.Warn(Warning{
			Message: "INVALID_OPTIONAL_DATE",
			Field:   "TRANSACTION.DTSTART",
			Value:   n.Value("DTSTART"),
		})
	}
	t.AvailableAt = parseOptionalOFXDate(n.Value("DTAVAIL"))
	if t.AvailableAt.Present && !t.AvailableAt.Valid {
		ctx.Warn(Warning{
			Message: "INVALID_OPTIONAL_DATE",
			Field:   "TRANSACTION.DTUSER",
			Value:   n.Value("DTUSER"),
		})
	}

	return t, nil
}

func nodeValues(n *sgml.Node) map[string]string {
	out := make(map[string]string)
	if n == nil {
		return out
	}
	for _, c := range n.Children {
		if len(c.Children) == 0 {
			out[c.Name] = strings.TrimSpace(c.Text)
		}
	}
	return out
}
