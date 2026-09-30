package ofx

import (
	"errors"
	"math/big"
	"time"
)

// Header represents the OFX file header that precedes the SGML/XML body.
type Header struct {
	OFXHeader   string
	Data        string
	Version     string
	Security    string
	Encoding    string
	Charset     string
	Compression string
	OldFileUID  string
	NewFileUID  string
	Extra       map[string]string
}

// Document is the normalized representation returned by Parse.
// Version-specific parsers populate this model so callers don't need to know
// whether a source file was SGML (1.x) or XML (2.x).
type Document struct {
	Header     Header
	Statements []BankStatement
	RawVersion string
	Profile    string
	Warnings   []Warning
}

// BankAccount identifies the account associated with a statement.
type BankAccount struct {
	BankID      string
	BranchID    string
	AccountID   string
	AccountType string
	AccountKey  string
}

// Balance is an account balance and the timestamp to which it applies.
type Balance struct {
	Amount string
	AsOf   OptionalDate
}

// BankStatement represents a BANKMSGSRSV1/STMTRS statement.
type BankStatement struct {
	Currency         string
	Account          BankAccount
	StartDate        OptionalDate
	EndDate          OptionalDate
	Transactions     []Transaction
	LedgerBalance    *Balance
	AvailableBalance *Balance
	Raw              map[string]string
}

type Warning struct {
	Code      string
	Field     string
	Value     string
	Message   string
	Statement int
	FITID     string
}

// Transaction is a normalized OFX STMTTRN record.
// Amount is kept as its exact decimal text to avoid binary floating point
// errors in financial reconciliation.
type Transaction struct {
	Type          string
	PostedAt      time.Time
	UserAt        OptionalDate
	AvailableAt   OptionalDate
	Amount        string
	FITID         string
	CorrectFITID  string
	CorrectAction string
	ServerID      string
	CheckNumber   string
	Reference     string
	SIC           string
	PayeeID       string
	Name          string
	ExtendedName  string
	Memo          string
	Raw           map[string]string
}

// AmountRat parses the transaction amount exactly as a rational number.
func (t Transaction) AmountRat() (*big.Rat, error) {
	if t.Amount == "" {
		return nil, errors.New("ofx: empty transaction amount")
	}
	r, ok := new(big.Rat).SetString(t.Amount)
	if !ok {
		return nil, errors.New("ofx: invalid transaction amount: " + t.Amount)
	}
	return r, nil
}

type OptionalDate struct {
	Time    time.Time
	Valid   bool
	Present bool
	Raw     string
}

func (d OptionalDate) IsZero() bool {
	return !d.Valid
}

func (d OptionalDate) Value() (time.Time, bool) {
	return d.Time, d.Valid
}
