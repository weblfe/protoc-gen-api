// Refer: https://github.com/grpc-ecosystem/grpc-gateway/blob/4ba7ec0bc390cae4a2d03625ac122aa8a772ac3a/protoc-gen-grpc-gateway/httprule/parse_test.go
package grammar_test

import (
	"flag"
	"fmt"
	"github.com/weblfe/protoc-gen-api/pkg/grammar"
	"reflect"
	"testing"

	"github.com/golang/glog"
)

func TestTokenize(t *testing.T) {
	for _, spec := range []struct {
		src    string
		tokens []string
	}{
		{
			src:    "",
			tokens: []string{grammar.EOF},
		},
		{
			src:    "v1",
			tokens: []string{"v1", grammar.EOF},
		},
		{
			src:    "v1/b",
			tokens: []string{"v1", "/", "b", grammar.EOF},
		},
		{
			src:    "v1/endpoint/*",
			tokens: []string{"v1", "/", "endpoint", "/", "*", grammar.EOF},
		},
		{
			src:    "v1/endpoint/**",
			tokens: []string{"v1", "/", "endpoint", "/", "**", grammar.EOF},
		},
		{
			src: "v1/b/{bucket_name=*}",
			tokens: []string{
				"v1", "/",
				"b", "/",
				"{", "bucket_name", "=", "*", "}",
				grammar.EOF,
			},
		},
		{
			src: "v1/b/{bucket_name=buckets/*}",
			tokens: []string{
				"v1", "/",
				"b", "/",
				"{", "bucket_name", "=", "buckets", "/", "*", "}",
				grammar.EOF,
			},
		},
		{
			src: "v1/b/{bucket_name=buckets/*}/o",
			tokens: []string{
				"v1", "/",
				"b", "/",
				"{", "bucket_name", "=", "buckets", "/", "*", "}", "/",
				"o",
				grammar.EOF,
			},
		},
		{
			src: "v1/b/{bucket_name=buckets/*}/o/{name}",
			tokens: []string{
				"v1", "/",
				"b", "/",
				"{", "bucket_name", "=", "buckets", "/", "*", "}", "/",
				"o", "/", "{", "name", "}",
				grammar.EOF,
			},
		},
		{
			src: "v1/a=b&c=d;e=f:g/endpoint.rdf",
			tokens: []string{
				"v1", "/",
				"a=b&c=d;e=f:g", "/",
				"endpoint.rdf",
				grammar.EOF,
			},
		},
	} {
		tokens, verb := grammar.Tokenize(spec.src)
		if got, want := tokens, spec.tokens; !reflect.DeepEqual(got, want) {
			t.Errorf("tokenize(%q) = %q, _; want %q, _", spec.src, got, want)
		}
		if got, want := verb, ""; got != want {
			t.Errorf("tokenize(%q) = _, %q; want _, %q", spec.src, got, want)
		}

		src := fmt.Sprintf("%s:%s", spec.src, "LOCK")
		tokens, verb = grammar.Tokenize(src)
		if got, want := tokens, spec.tokens; !reflect.DeepEqual(got, want) {
			t.Errorf("tokenize(%q) = %q, _; want %q, _", src, got, want)
		}
		if got, want := verb, "LOCK"; got != want {
			t.Errorf("tokenize(%q) = _, %q; want _, %q", src, got, want)
		}
	}
}

func TestParseSegments(t *testing.T) {
	flag.Set("v", "3")
	for _, spec := range []struct {
		tokens []string
		want   []grammar.Segment
	}{
		{
			tokens: []string{"v1", grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("v1"),
			},
		},
		{
			tokens: []string{"/", grammar.EOF},
			want: []grammar.Segment{
				grammar.Wildcard{},
			},
		},
		{
			tokens: []string{"-._~!$&'()*+,;=:@", grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("-._~!$&'()*+,;=:@"),
			},
		},
		{
			tokens: []string{"%e7%ac%ac%e4%b8%80%e7%89%88", grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("%e7%ac%ac%e4%b8%80%e7%89%88"),
			},
		},
		{
			tokens: []string{"v1", "/", "*", grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("v1"),
				grammar.Wildcard{},
			},
		},
		{
			tokens: []string{"v1", "/", "**", grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("v1"),
				grammar.DeepWildcard{},
			},
		},
		{
			tokens: []string{"{", "name", "}", grammar.EOF},
			want: []grammar.Segment{
				grammar.Variable{
					Path: "name",
					Segments: []grammar.Segment{
						grammar.Wildcard{},
					},
				},
			},
		},
		{
			tokens: []string{"{", "name", "=", "*", "}", grammar.EOF},
			want: []grammar.Segment{
				grammar.Variable{
					Path: "name",
					Segments: []grammar.Segment{
						grammar.Wildcard{},
					},
				},
			},
		},
		{
			tokens: []string{"{", "field", ".", "nested", ".", "nested2", "=", "*", "}", grammar.EOF},
			want: []grammar.Segment{
				grammar.Variable{
					Path: "field.nested.nested2",
					Segments: []grammar.Segment{
						grammar.Wildcard{},
					},
				},
			},
		},
		{
			tokens: []string{"{", "name", "=", "a", "/", "b", "/", "*", "}", grammar.EOF},
			want: []grammar.Segment{
				grammar.Variable{
					Path: "name",
					Segments: []grammar.Segment{
						grammar.Literal("a"),
						grammar.Literal("b"),
						grammar.Wildcard{},
					},
				},
			},
		},
		{
			tokens: []string{
				"v1", "/",
				"{",
				"name", ".", "nested", ".", "nested2",
				"=",
				"a", "/", "b", "/", "*",
				"}", "/",
				"o", "/",
				"{",
				"another_name",
				"=",
				"a", "/", "b", "/", "*", "/", "c",
				"}", "/",
				"**",
				grammar.EOF},
			want: []grammar.Segment{
				grammar.Literal("v1"),
				grammar.Variable{
					Path: "name.nested.nested2",
					Segments: []grammar.Segment{
						grammar.Literal("a"),
						grammar.Literal("b"),
						grammar.Wildcard{},
					},
				},
				grammar.Literal("o"),
				grammar.Variable{
					Path: "another_name",
					Segments: []grammar.Segment{
						grammar.Literal("a"),
						grammar.Literal("b"),
						grammar.Wildcard{},
						grammar.Literal("c"),
					},
				},
				grammar.DeepWildcard{},
			},
		},
	} {
		p := grammar.NewParser(grammar.ApplyTokens(spec.tokens...))
		segs, err := p.TopLevelSegments()
		if err != nil {
			t.Errorf("parser{%q}.Segments() failed with %v; want success", spec.tokens, err)
			continue
		}
		if got, want := segs, spec.want; !reflect.DeepEqual(got, want) {
			t.Errorf("parser{%q}.Segments() = %#v; want %#v", spec.tokens, got, want)
		}
		if got := p.GetTokens(); len(got) > 0 {
			t.Errorf("p.tokens = %q; want []; spec.tokens=%q", got, spec.tokens)
		}
	}
}

func TestParseSegmentsWithErrors(t *testing.T) {
	flag.Set("v", "3")
	for _, spec := range []struct {
		tokens []string
	}{
		{
			// double slash
			tokens: []string{"//", grammar.EOF},
		},
		{
			// invalid Literal
			tokens: []string{"a?b", grammar.EOF},
		},
		{
			// invalid percent-encoding
			tokens: []string{"%", grammar.EOF},
		},
		{
			// invalid percent-encoding
			tokens: []string{"%2", grammar.EOF},
		},
		{
			// invalid percent-encoding
			tokens: []string{"a%2z", grammar.EOF},
		},
		{
			// empty Segments
			tokens: []string{grammar.EOF},
		},
		{
			// unterminated Variable
			tokens: []string{"{", "name", grammar.EOF},
		},
		{
			// unterminated Variable
			tokens: []string{"{", "name", "=", grammar.EOF},
		},
		{
			// unterminated Variable
			tokens: []string{"{", "name", "=", "*", grammar.EOF},
		},
		{
			// empty component in field Path
			tokens: []string{"{", "name", ".", "}", grammar.EOF},
		},
		{
			// empty component in field Path
			tokens: []string{"{", "name", ".", ".", "nested", "}", grammar.EOF},
		},
		{
			// invalid character in identifier
			tokens: []string{"{", "field-name", "}", grammar.EOF},
		},
		{
			// no slash between Segments
			tokens: []string{"v1", "endpoint", grammar.EOF},
		},
		{
			// no slash between Segments
			tokens: []string{"v1", "{", "name", "}", grammar.EOF},
		},
	} {

		p := grammar.NewParser(grammar.ApplyTokens(spec.tokens...))
		segs, err := p.TopLevelSegments()
		if err == nil {
			t.Errorf("parser{%q}.Segments() succeeded; want InvalidTemplateError; accepted %#v", spec.tokens, segs)
			continue
		}
		glog.V(1).Info(err)
	}
}
