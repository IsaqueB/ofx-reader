package ofx

import "testing"

func TestParseOFXDate(t *testing.T) {
	d, err := parseOFXDate("20260915103000.123[-3:BRT]")
	if err != nil {
		t.Fatal(err)
	}
	_, off := d.Zone()
	if off != -3*3600 {
		t.Fatalf("offset=%d", off)
	}
	if d.Nanosecond() != 123000000 {
		t.Fatalf("nanos=%d", d.Nanosecond())
	}
}
