package expr

import (
	"github.com/brimdata/zed"
	"github.com/brimdata/zed/zcode"
)

// Context is an interface to a scope and value allocator for expressions.
// This allows external packages to implement efficient temporary buffering
// of Zed values both for let-style temporary variables accessible via
// the scope and for allocating results.
type Context interface {
	// Vars() accesses the variables reachable in the current scope.
	Vars() []zed.Value
	//XXX there should be two NewValues: one when bytes is already inside
	// of the context... another when you need to copy those bytes into
	// this context.
	zed.Allocator
}

type allocator struct{}

var _ Context = (*allocator)(nil)

func NewContext() *allocator {
	return &allocator{}
}

func (*allocator) NewValue(typ zed.Type, bytes zcode.Bytes) *zed.Value {
	return zed.NewValue(typ, bytes)
}

func (*allocator) CopyValue(val *zed.Value) *zed.Value { return val.Copy() }

func (*allocator) Vars() []zed.Value {
	return nil
}

type ResetContext struct {
	buf  []byte
	vals []zed.Value
	vars []zed.Value
}

var _ Context = (*ResetContext)(nil)

func (r *ResetContext) NewValue(typ zed.Type, b zcode.Bytes) *zed.Value {
	// Preserve b if nil or empty.
	if len(b) > 0 {
		n := len(r.buf)
		r.buf = append(r.buf, b...)
		b = r.buf[n:]
	}
	r.vals = append(r.vals, *zed.NewValue(typ, b))
	return &r.vals[len(r.vals)-1]
}

func (r *ResetContext) CopyValue(val *zed.Value) *zed.Value {
	val2 := r.NewValue(nil, nil)
	*val2 = *val
	return val2
}

func (r *ResetContext) Reset() *ResetContext {
	r.buf = r.buf[:0]
	r.vals = r.vals[:0]
	return r
}

func (r *ResetContext) SetVars(vars []zed.Value) { r.vars = vars }
func (r *ResetContext) Vars() []zed.Value        { return r.vars }
