package grammar

import (
		"fmt"
		"strings"
)

// Refer: https://github.com/grpc-ecosystem/grpc-gateway/blob/4ba7ec0bc390cae4a2d03625ac122aa8a772ac3a/protoc-gen-grpc-gateway/httprule/types.go


type Segment interface {
		fmt.Stringer
}

type Wildcard struct{}

type DeepWildcard struct{}

type Literal string

type Variable struct {
		Path     string
		Segments []Segment
}

func (Wildcard) String() string {
		return "*"
}

func (DeepWildcard) String() string {
		return "**"
}

func (l Literal) String() string {
		return string(l)
}

func (v Variable) String() string {
		var segments []string
		for _, s := range v.Segments {
				segments = append(segments, s.String())
		}
		return fmt.Sprintf("{%s=%s}", v.Path, strings.Join(segments, "/"))
}

func (v Variable)GetPath() string  {
		return v.Path
}

func (v Variable)GetSegments() []Segment  {
		return v.Segments
}