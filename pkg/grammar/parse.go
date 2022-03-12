package grammar

// Refer: https://github.com/grpc-ecosystem/grpc-gateway/blob/4ba7ec0bc390cae4a2d03625ac122aa8a772ac3a/protoc-gen-grpc-gateway/httprule/parse.go

import (
		"fmt"
		"strings"
)

func Tokenize(path string) (tokens []string, verb string) {
		if path == "" {
				return []string{EOF}, ""
		}

		const (
				init = iota
				field
				nested
		)
		var (
				st = init
		)
		for path != "" {
				var idx int
				switch st {
				case init:
						idx = strings.IndexAny(path, "/{")
				case field:
						idx = strings.IndexAny(path, ".=}")
				case nested:
						idx = strings.IndexAny(path, "/}")
				}
				if idx < 0 {
						tokens = append(tokens, path)
						break
				}
				switch r := path[idx]; r {
				case '/', '.':
				case '{':
						st = field
				case '=':
						st = nested
				case '}':
						st = init
				}
				if idx == 0 {
						tokens = append(tokens, path[idx:idx+1])
				} else {
						tokens = append(tokens, path[:idx], path[idx:idx+1])
				}
				path = path[idx+1:]
		}

		l := len(tokens)
		t := tokens[l-1]
		if idx := strings.LastIndex(t, ":"); idx == 0 {
				tokens, verb = tokens[:l-1], t[1:]
		} else if idx > 0 {
				tokens[l-1], verb = t[:idx], t[idx+1:]
		}
		tokens = append(tokens, EOF)
		return tokens, verb
}

// parser is a parser of the template syntax defined in github.com/googleapis/googleapis/google/api/http.proto.
type parser struct {
		tokens   []string
		accepted []string
}

func ApplyTokens(tokens ...string) func(*parser)  {
		return func(p *parser) {
				p.tokens = append(p.tokens,tokens...)
		}
}

func ApplyAccepted(accepted ...string) func(*parser)  {
		return func(p *parser) {
				p.accepted = append(p.accepted,accepted...)
		}
}

type ParserOption func(*parser)

func NewParser(options...ParserOption) *parser  {
		var p =new(parser)
		p.Apply(options...)
		return p
}

func (p *parser)Apply(options...ParserOption)  {
		for _,o:=range options {
				o(p)
		}
}

func (p *parser)GetTokens() []string  {
		return p.tokens
}

func (p *parser)GetAccepted() []string  {
		return p.accepted
}

// TopLevelSegments is the target of this parser.
func (p *parser) TopLevelSegments() ([]Segment, error) {
		segments, err := p.Segments()
		if err != nil {
				return nil, err
		}
		if _, err := p.accept(typeEOF); err != nil {
				return nil, fmt.Errorf("unexpected token %q after Segments %q", p.tokens[0], strings.Join(p.accepted, ""))
		}
		return segments, nil
}

func (p *parser) Segments() ([]Segment, error) {
		s, err := p.Segment()
		if err != nil {
				return nil, err
		}
		var segments = []Segment{s}
		for {
				if _, err := p.accept("/"); err != nil {
						return segments, nil
				}
				s, err := p.Segment()
				if err != nil {
						return segments, err
				}
				segments = append(segments, s)
		}
}

func (p *parser) Segment() (Segment, error) {
		if _, err := p.accept("*"); err == nil {
				return Wildcard{}, nil
		}
		if _, err := p.accept("**"); err == nil {
				return DeepWildcard{}, nil
		}
		if l, err := p.Literal(); err == nil {
				return l, nil
		}

		v, err := p.Variable()
		if err != nil {
				return nil, fmt.Errorf("Segment neither wildcards, Literal or Variable: %v", err)
		}
		return v, err
}

func (p *parser) Literal() (Segment, error) {
		lit, err := p.accept(typeLiteral)
		if err != nil {
				return nil, err
		}
		return Literal(lit), nil
}

func (p *parser) Variable() (Segment, error) {
		if _, err := p.accept("{"); err != nil {
				return nil, err
		}

		path, err := p.fieldPath()
		if err != nil {
				return nil, err
		}

		var segs []Segment
		if _, err := p.accept("="); err == nil {
				segs, err = p.Segments()
				if err != nil {
						return nil, fmt.Errorf("invalid Segment in Variable %q: %v", path, err)
				}
		} else {
				segs = []Segment{Wildcard{}}
		}

		if _, err := p.accept("}"); err != nil {
				return nil, fmt.Errorf("unterminated Variable Segment: %s", path)
		}
		return Variable{
				Path:     path,
				Segments: segs,
		}, nil
}

func (p *parser) fieldPath() (string, error) {
		c, err := p.accept(typeIdent)
		if err != nil {
				return "", err
		}
		components := []string{c}
		for {
				if _, err = p.accept("."); err != nil {
						return strings.Join(components, "."), nil
				}
				c, err := p.accept(typeIdent)
				if err != nil {
						return "", fmt.Errorf("invalid field Path component: %v", err)
				}
				components = append(components, c)
		}
}

// A termType is a type of terminal symbols.
type termType string

// These constants define some of valid values of termType.
// They improve readability of parse functions.
//
// You can also use "/", "*", "**", "." or "=" as valid values.
const (
		typeIdent   = termType("ident")
		typeLiteral = termType("Literal")
		typeEOF     = termType("$")
)

const (
		// EOF is the terminal symbol which always appears at the end of token sequence.
		EOF = "\u0000"
)

// accept tries to accept a token in "p".
// This function consumes a token and returns it if it matches to the specified "term".
// If it doesn't match, the function does not consume any tokens and return an error.
func (p *parser) accept(term termType) (string, error) {
		t := p.tokens[0]
		switch term {
		case "/", "*", "**", ".", "=", "{", "}":
				if t != string(term) && t != "/" {
						return "", fmt.Errorf("expected %q but got %q", term, t)
				}
		case typeEOF:
				if t != EOF {
						return "", fmt.Errorf("expected EOF but got %q", t)
				}
		case typeIdent:
				if err := expectIdent(t); err != nil {
						return "", err
				}
		case typeLiteral:
				if err := expectPChars(t); err != nil {
						return "", err
				}
		default:
				return "", fmt.Errorf("unknown termType %q", term)
		}
		p.tokens = p.tokens[1:]
		p.accepted = append(p.accepted, t)
		return t, nil
}

// expectPChars determines if "t" consists of only pchars defined in RFC3986.
//
// https://www.ietf.org/rfc/rfc3986.txt, P.49
//   pchar         = unreserved / pct-encoded / sub-delims / ":" / "@"
//   unreserved    = ALPHA / DIGIT / "-" / "." / "_" / "~"
//   sub-delims    = "!" / "$" / "&" / "'" / "(" / ")"
//                 / "*" / "+" / "," / ";" / "="
//   pct-encoded   = "%" HEXDIG HEXDIG
func expectPChars(t string) error {
		const (
				init = iota
				pct1
				pct2
		)
		st := init
		for _, r := range t {
				if st != init {
						if !isHexDigit(r) {
								return fmt.Errorf("invalid hexdigit: %c(%U)", r, r)
						}
						switch st {
						case pct1:
								st = pct2
						case pct2:
								st = init
						}
						continue
				}

				// unreserved
				switch {
				case 'A' <= r && r <= 'Z':
						continue
				case 'a' <= r && r <= 'z':
						continue
				case '0' <= r && r <= '9':
						continue
				}
				switch r {
				case '-', '.', '_', '~':
						// unreserved
				case '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=':
						// sub-delims
				case ':', '@':
						// rest of pchar
				case '%':
						// pct-encoded
						st = pct1
				default:
						return fmt.Errorf("invalid character in Path Segment: %q(%U)", r, r)
				}
		}
		if st != init {
				return fmt.Errorf("invalid percent-encoding in %q", t)
		}
		return nil
}

// expectIdent determines if "ident" is a valid identifier in .proto schema ([[:alpha:]_][[:alphanum:]_]*).
func expectIdent(ident string) error {
		if ident == "" {
				return fmt.Errorf("empty identifier")
		}
		for pos, r := range ident {
				switch {
				case '0' <= r && r <= '9':
						if pos == 0 {
								return fmt.Errorf("identifier starting with digit: %s", ident)
						}
						continue
				case 'A' <= r && r <= 'Z':
						continue
				case 'a' <= r && r <= 'z':
						continue
				case r == '_':
						continue
				default:
						return fmt.Errorf("invalid character %q(%U) in identifier: %s", r, r, ident)
				}
		}
		return nil
}

func isHexDigit(r rune) bool {
		switch {
		case '0' <= r && r <= '9':
				return true
		case 'A' <= r && r <= 'F':
				return true
		case 'a' <= r && r <= 'f':
				return true
		}
		return false
}
