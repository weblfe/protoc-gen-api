package generators

import (
	"fmt"
	"github.com/weblfe/protoc-gen-api/pkg/grammar"
	"regexp"
	"sort"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var toCamelCaseRe = regexp.MustCompile(`(^[A-Za-z])|(_|\.)([A-Za-z])`)

func toCamelCase(str string) string {
	return toCamelCaseRe.ReplaceAllStringFunc(str, func(s string) string {
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

type PathParam struct {
	Index  int
	Name   string
	GoName string
}

type PathParamArr []*PathParam

func (p PathParamArr) Len() int {
	return len(p)
}

func (p PathParamArr) Less(i, j int) bool {
	var src, dst = p[i], p[j]
	if len(strings.Split(src.Name, ".")) < len(strings.Split(dst.Name, ".")) {
		return true
	}
	return src.Name < dst.Name
}

func (p PathParamArr) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (t *PathParam) GetGoNamesWithSplit() []string {
	names := make([]string, 0)
	ps := strings.Split(t.GoName, ".")
	for i := range ps {
		if i == 0 {
			continue
		}
		names = append(names, strings.Join(ps[0:i], "."))
	}
	return names
}

func parsePathParam(pattern string) ([]*PathParam, error) {
	if !strings.HasPrefix(pattern, "/") {
		return nil, fmt.Errorf("no leading /")
	}
	var (
		err       error
		segments  []grammar.Segment
		tokens, _ = grammar.Tokenize(pattern[1:])
	)
	p := grammar.NewParser(grammar.ApplyTokens(tokens...))
	if segments, err = p.TopLevelSegments(); err != nil {
		return nil, err
	}
	var params PathParamArr
	for i, seg := range segments {
		switch seg.(type) {
		case grammar.Variable:
			var v = seg.(grammar.Variable)
			params = append(params, &PathParam{
				Index:  i + 1,
				Name:   v.GetPath(),
				GoName: toCamelCase(v.GetPath()),
			})
		case *grammar.Variable:
			var v = seg.(*grammar.Variable)
			params = append(params, &PathParam{
				Index:  i + 1,
				Name:   v.GetPath(),
				GoName: toCamelCase(v.GetPath()),
			})
		}
	}
	sort.Sort(&params)
	return params, nil
}

type queryParam struct {
	*protogen.Field

	GoName string
	Name   string
}

func createQueryParams(method *protogen.Method) []*queryParam {
	queryParams := make([]*queryParam, 0)

	var f func(parent *queryParam, fields []*protogen.Field)

	f = func(parent *queryParam, fields []*protogen.Field) {
		for _, field := range fields {
			if field.Desc.Kind() == protoreflect.MessageKind {
				q := &queryParam{
					Field:  field,
					GoName: fmt.Sprintf("%s.", field.GoName),
					Name:   fmt.Sprintf("%s.", field.Desc.Name()),
				}
				f(q, field.Message.Fields)
				continue
			}
			queryParams = append(queryParams, &queryParam{
				Field:  field,
				GoName: fmt.Sprintf("%s%s", parent.GoName, field.GoName),
				Name:   fmt.Sprintf("%s%s", parent.Name, field.Desc.Name()),
			})
		}
	}

	f(&queryParam{GoName: "", Name: ""}, method.Input.Fields)

	return queryParams
}
