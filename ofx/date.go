package ofx

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseOFXDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}

	tzOffsetSeconds := 0
	hasTZ := false
	if i := strings.IndexByte(s, '['); i >= 0 {
		if j := strings.IndexByte(s[i:], ']'); j >= 0 {
			tz := s[i+1 : i+j]
			offset := tz
			if k := strings.IndexByte(tz, ':'); k >= 0 {
				offset = tz[:k]
			}
			if f, err := strconv.ParseFloat(strings.TrimSpace(offset), 64); err == nil {
				tzOffsetSeconds = int(f * 3600)
				hasTZ = true
			}
			s = s[:i]
		}
	}

	fracNanos := 0
	if i := strings.IndexByte(s, '.'); i >= 0 {
		frac := s[i+1:]
		s = s[:i]
		if len(frac) > 9 {
			frac = frac[:9]
		}
		frac += strings.Repeat("0", 9-len(frac))
		if n, err := strconv.Atoi(frac); err == nil {
			fracNanos = n
		}
	}

	if len(s) < 8 {
		return time.Time{}, fmt.Errorf("invalid OFX date %q", s)
	}
	layouts := []string{"20060102150405", "200601021504", "2006010215", "20060102"}
	var t time.Time
	var err error
	for _, layout := range layouts {
		if len(s) != len(layout) {
			continue
		}
		loc := time.Local
		if hasTZ {
			loc = time.FixedZone("OFX", tzOffsetSeconds)
		}
		t, err = time.ParseInLocation(layout, s, loc)
		if err == nil {
			return t.Add(time.Duration(fracNanos)), nil
		}
	}
	if err == nil {
		err = fmt.Errorf("unsupported OFX date length")
	}
	return time.Time{}, fmt.Errorf("invalid OFX date %q: %w", s, err)
}

func parseOptionalOFXDate(raw string) OptionalDate {
	raw = strings.TrimSpace(raw)

	if raw == "" {
		return OptionalDate{}
	}

	result := OptionalDate{
		Raw:     raw,
		Present: true,
	}

	t, err := parseOFXDate(raw)
	if err != nil {
		return result
	}

	result.Time = t
	result.Valid = true
	return result
}
