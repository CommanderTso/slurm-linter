package parser_test

import (
	"reflect"
	"testing"

	"github.com/CommanderTso/slurm-linter/internal/parser"
)

func TestExpandNodeRange(t *testing.T) {
	cases := []struct {
		expr string
		want []string
		err  bool
	}{
		// plain name, no brackets
		{"node01", []string{"node01"}, false},
		// single number in brackets
		{"node[5]", []string{"node5"}, false},
		// simple range
		{"node[01-03]", []string{"node01", "node02", "node03"}, false},
		// zero-padded range
		{"node[001-003]", []string{"node001", "node002", "node003"}, false},
		// comma list
		{"node[1,3,5]", []string{"node1", "node3", "node5"}, false},
		// mixed range and singles
		{"node[01-03,05]", []string{"node01", "node02", "node03", "node05"}, false},
		// multi-range
		{"node[01-02,04-05]", []string{"node01", "node02", "node04", "node05"}, false},
		// missing close bracket → error
		{"node[01-03", nil, true},
		// start > end → error
		{"node[05-03]", nil, true},
	}

	for _, c := range cases {
		got, err := parser.ExpandNodeRange(c.expr)
		if c.err {
			if err == nil {
				t.Errorf("ExpandNodeRange(%q): expected error, got nil", c.expr)
			}
			continue
		}
		if err != nil {
			t.Errorf("ExpandNodeRange(%q): unexpected error: %v", c.expr, err)
			continue
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("ExpandNodeRange(%q) = %v, want %v", c.expr, got, c.want)
		}
	}
}
