package ofx

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// VersionParser is implemented by every supported OFX version family.
type VersionParser interface {
	Version() string
	Parse(header Header, body []byte) (*Document, error)
}

// Parser is the reusable OFX parser. It is safe for concurrent Parse calls
// after construction, provided callers do not mutate its registries.
type Parser struct {
	versions map[string]VersionParser
	profiles *ProfileRegistry
}

func NewParser() *Parser {
	p := &Parser{
		versions: make(map[string]VersionParser),
		profiles: NewProfileRegistry(),
	}
	p.RegisterVersion(newV102Parser())
	return p
}

func (p *Parser) RegisterVersion(v VersionParser) {
	if p == nil || v == nil {
		return
	}
	p.versions[v.Version()] = v
}

func (p *Parser) RegisterProfile(profile BankProfile) {
	if p == nil {
		return
	}
	p.profiles.Register(profile)
}

// Parse reads an OFX document from r.
func (p *Parser) Parse(r io.Reader) (*Document, error) {
	if p == nil {
		return nil, errors.New("ofx: nil parser")
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("ofx: read: %w", err)
	}
	return p.ParseBytes(data)
}

// ParseBytes parses an OFX document already loaded in memory.
func (p *Parser) ParseBytes(data []byte) (*Document, error) {
	header, body, err := splitHeader(data)
	if err != nil {
		return nil, err
	}
	impl := p.versions[header.Version]
	if impl == nil {
		return nil, &UnsupportedVersionError{Version: header.Version}
	}
	doc, err := impl.Parse(header, body)
	if err != nil {
		return nil, err
	}

	var profileName string
	for i := range doc.Statements {
		name, err := p.profiles.apply(header, &doc.Statements[i])
		if err != nil {
			return nil, fmt.Errorf("ofx: bank profile %q: %w", name, err)
		}
		if profileName == "" && name != "" {
			profileName = name
		}
	}
	doc.Profile = profileName
	return doc, nil
}

// ParseFile is a convenience helper for filesystem-based applications.
func (p *Parser) ParseFile(path string) (*Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("ofx: open %s: %w", path, err)
	}
	defer f.Close()
	return p.Parse(f)
}

// Parse uses a new default parser supporting the built-in versions.
func Parse(r io.Reader) (*Document, error)      { return NewParser().Parse(r) }
func ParseBytes(data []byte) (*Document, error) { return NewParser().ParseBytes(data) }
func ParseFile(path string) (*Document, error)  { return NewParser().ParseFile(path) }

type UnsupportedVersionError struct{ Version string }

func (e *UnsupportedVersionError) Error() string {
	return "ofx: unsupported OFX version " + e.Version
}

func splitHeader(data []byte) (Header, []byte, error) {
	idx := bytes.Index(bytes.ToUpper(data), []byte("<OFX"))
	if idx < 0 {
		return Header{}, nil, errors.New("ofx: <OFX> root not found")
	}

	rawHeader := data[:idx]
	body := data[idx:]
	h := Header{Extra: make(map[string]string)}

	scanner := bufio.NewScanner(bytes.NewReader(rawHeader))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		key = strings.ToUpper(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		switch key {
		case "OFXHEADER":
			h.OFXHeader = value
		case "DATA":
			h.Data = value
		case "VERSION":
			h.Version = value
		case "SECURITY":
			h.Security = value
		case "ENCODING":
			h.Encoding = value
		case "CHARSET":
			h.Charset = value
		case "COMPRESSION":
			h.Compression = value
		case "OLDFILEUID":
			h.OldFileUID = value
		case "NEWFILEUID":
			h.NewFileUID = value
		default:
			h.Extra[key] = value
		}
	}
	if err := scanner.Err(); err != nil {
		return Header{}, nil, fmt.Errorf("ofx: scan header: %w", err)
	}
	if h.Version == "" {
		return Header{}, nil, errors.New("ofx: VERSION missing from header")
	}
	return h, body, nil
}
