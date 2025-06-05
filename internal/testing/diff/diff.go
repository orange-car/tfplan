package diff

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Check(t *testing.T, expect any, got any, opts ...cmp.Option) {
	d := cmp.Diff(expect, got, opts...)
	if d != "" {
		t.Errorf(`

### %s failed

Diff:
-------------
%s
-------------

`, t.Name(), d)
	}
}
