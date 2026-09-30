package textcodec

import (
	"bytes"
	"strings"
	"unicode/utf8"
)

// Decode converts common OFX 1.x encodings to UTF-8 using only the Go standard
// library. Windows-1252 is implemented locally because x/text is intentionally
// not a dependency of this project.
func Decode(data []byte, encoding, charset string) string {
	enc := strings.ToUpper(strings.TrimSpace(encoding))
	cs := strings.ToUpper(strings.TrimSpace(charset))

	if enc == "UNICODE" || enc == "UTF-8" || cs == "UTF-8" || cs == "65001" {
		if utf8.Valid(data) {
			return string(data)
		}
	}
	if cs == "1252" || cs == "WINDOWS-1252" || cs == "CP1252" {
		return decodeWindows1252(data)
	}
	if cs == "8859-1" || cs == "ISO-8859-1" || cs == "ISO8859-1" {
		var b strings.Builder
		for _, c := range data {
			b.WriteRune(rune(c))
		}
		return b.String()
	}
	if utf8.Valid(data) {
		return string(data)
	}
	// OFX files frequently declare USASCII while actually containing bytes in
	// Windows-1252. Falling back to 1252 is more useful than emitting U+FFFD.
	return decodeWindows1252(data)
}

func decodeWindows1252(data []byte) string {
	var out bytes.Buffer
	table := map[byte]rune{
		0x80: '€', 0x82: '‚', 0x83: 'ƒ', 0x84: '„', 0x85: '…', 0x86: '†', 0x87: '‡',
		0x88: 'ˆ', 0x89: '‰', 0x8A: 'Š', 0x8B: '‹', 0x8C: 'Œ', 0x8E: 'Ž', 0x91: '‘',
		0x92: '’', 0x93: '“', 0x94: '”', 0x95: '•', 0x96: '–', 0x97: '—', 0x98: '˜',
		0x99: '™', 0x9A: 'š', 0x9B: '›', 0x9C: 'œ', 0x9E: 'ž', 0x9F: 'Ÿ',
	}
	for _, c := range data {
		if r, ok := table[c]; ok {
			out.WriteRune(r)
			continue
		}
		out.WriteRune(rune(c))
	}
	return out.String()
}
