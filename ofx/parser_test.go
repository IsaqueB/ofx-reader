package ofx

import (
	"errors"
	"os"
	"testing"
)

func TestParseV102(t *testing.T) {
	doc, err := ParseFile("../testdata/sample_102.ofx")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Header.Version != "102" {
		t.Fatalf("version=%q", doc.Header.Version)
	}
	if doc.RawVersion != "1.0.2" {
		t.Fatalf("raw version=%q", doc.RawVersion)
	}
	if len(doc.Statements) != 1 {
		t.Fatalf("statements=%d", len(doc.Statements))
	}
	st := doc.Statements[0]
	if st.Account.BankID != "001" {
		t.Fatalf("bank=%q", st.Account.BankID)
	}
	if st.Account.BranchID != "1234" {
		t.Fatalf("branch=%q", st.Account.BranchID)
	}
	if len(st.Transactions) != 2 {
		t.Fatalf("transactions=%d", len(st.Transactions))
	}
	if st.Transactions[0].Amount != "1234.56" {
		t.Fatalf("amount=%q", st.Transactions[0].Amount)
	}
	if st.Transactions[1].FITID != "001-20260916-2" {
		t.Fatalf("fitid=%q", st.Transactions[1].FITID)
	}
	if st.LedgerBalance == nil || st.LedgerBalance.Amount != "1134.66" {
		t.Fatalf("ledger=%+v", st.LedgerBalance)
	}
}

func TestAmountRatIsExact(t *testing.T) {
	r, err := (Transaction{Amount: "0.10"}).AmountRat()
	if err != nil {
		t.Fatal(err)
	}
	if r.RatString() != "1/10" {
		t.Fatalf("got %s", r.RatString())
	}
}

func TestUnsupportedVersion(t *testing.T) {
	data, err := os.ReadFile("../testdata/sample_102.ofx")
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i+11 <= len(data); i++ {
		if string(data[i:i+11]) == "VERSION:102" {
			copy(data[i:i+11], []byte("VERSION:200"))
			break
		}
	}
	_, err = ParseBytes(data)
	var uv *UnsupportedVersionError
	if !errors.As(err, &uv) {
		t.Fatalf("expected UnsupportedVersionError, got %v", err)
	}
}

type testProfile struct{}

func (testProfile) Name() string                         { return "test" }
func (testProfile) Match(_ Header, s BankStatement) bool { return s.Account.BankID == "001" }
func (testProfile) Normalize(s *BankStatement) error     { s.Account.BankID = "001-normalized"; return nil }

func TestProfileExtensionPoint(t *testing.T) {
	p := NewParser()
	p.RegisterProfile(testProfile{})
	doc, err := p.ParseFile("../testdata/sample_102.ofx")
	if err != nil {
		t.Fatal(err)
	}
	if doc.Profile != "test" {
		t.Fatalf("profile=%q", doc.Profile)
	}
	if doc.Statements[0].Account.BankID != "001-normalized" {
		t.Fatal("profile not applied")
	}
}
