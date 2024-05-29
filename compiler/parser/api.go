package parser

import (
	"fmt"
	"strings"

	"github.com/brimdata/zed/compiler/ast"
)

// ParseZed calls ConcatSource followed by Parse.  If Parse returns an error,
// ConcatSource tries to convert it to an ErrorList.
func ParseZed(filenames []string, src string) (ast.Seq, *SourceSet, error) {
	sset, err := ConcatSource(filenames, src)
	if err != nil {
		return nil, nil, err
	}
	p, err := Parse("", sset.Contents)
	if err != nil {
		return nil, nil, convertErrList(err, sset)
	}
	return sliceOf[ast.Op](p), sset, nil
}

func convertErrList(err error, sset *SourceSet) error {
	errs, ok := err.(errList)
	if !ok {
		return err
	}
	var out ErrorList
	for _, e := range errs {
		pe, ok := e.(*parserError)
		if !ok {
			return err
		}
		out.Append("error parsing Zed", pe.pos.offset, -1)
	}
	out.AddSourceSet(sset)
	return out
}

type ErrorList []*Error

func (e *ErrorList) Append(msg string, pos, end int) {
	*e = append(*e, &Error{msg, pos, end, nil})
}

func (e ErrorList) AddSourceSet(sset *SourceSet) {
	for i := range e {
		e[i].sset = sset
	}
}

func (e ErrorList) Error() string {
	var b strings.Builder
	for i, err := range e {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

type Error struct {
	Msg  string
	Pos  int
	End  int
	sset *SourceSet
}

func (e *Error) Error() string {
	if e.sset == nil {
		return e.Msg
	}
	src := e.sset.SourceOf(e.Pos)
	start := src.Position(e.Pos)
	end := src.Position(e.End)
	var b strings.Builder
	fmt.Fprintf(&b, "%s (", e.Msg)
	if src.Filename != "" {
		fmt.Fprintf(&b, "%s: ", src.Filename)
	}
	line := src.LineOfPos(e.sset.Contents, e.Pos)
	fmt.Fprintf(&b, "line %d, column %d):\n%s\n", start.Line, start.Column, line)
	if end.IsValid() {
		formatSpanError(&b, line, start, end)
	} else {
		formatPointError(&b, start)
	}
	return b.String()
}

func formatSpanError(b *strings.Builder, line string, start, end Position) {
	col := start.Column - 1
	b.WriteString(strings.Repeat(" ", col))
	n := len(line) - col
	if start.Line == end.Line {
		n = end.Column - col
	}
	b.WriteString(strings.Repeat("~", n))
}

func formatPointError(b *strings.Builder, start Position) {
	col := start.Column - 1
	for k := 0; k < col; k++ {
		if k >= col-4 && k != col-1 {
			b.WriteByte('=')
		} else {
			b.WriteByte(' ')
		}
	}
	b.WriteString("^ ===")
}
