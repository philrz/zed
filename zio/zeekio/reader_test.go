package zeekio

import (
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/brimdata/zed"
	"github.com/brimdata/zed/pkg/nano"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReaderCRLF(t *testing.T) {
	arena := zed.NewArena()
	defer arena.Unref()

	input := `
#separator \x09
#set_separator	,
#empty_field	(empty)
#unset_field	-
#path	a
#fields	ts	i
#types	time	int
10.000000	1
`
	input = strings.ReplaceAll(input, "\n", "\r\n")
	r := NewReader(zed.NewContext(), strings.NewReader(input))
	defer runtime.KeepAlive(r)
	rec, err := r.Read()
	require.NoError(t, err)
	ts := rec.Deref(arena, "ts").AsTime()
	assert.Exactly(t, 10*nano.Ts(time.Second), ts)
	d := rec.Deref(arena, "i").AsInt()
	assert.Exactly(t, int64(1), d)
	rec, err = r.Read()
	require.NoError(t, err)
	assert.Nil(t, rec)
}
