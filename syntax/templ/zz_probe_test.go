package templ

import (
	"testing"

	templparser "github.com/a-h/templ/parser/v2"
)

func TestProbeSpread(t *testing.T) {
	src := "package main\n\ntempl A(user User) {\n\t<div { user.Attrs... }>x</div>\n}\n"
	tf, err := templparser.ParseString(src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	for _, decl := range tf.Nodes {
		c, ok := decl.(*templparser.TemplElement)
		if !ok {
			t.Logf("decl %T", decl)
			continue
		}
		t.Logf("element %s attrs=%d", c.Name, len(c.Attributes))
		for i, a := range c.Attributes {
			t.Logf("attr[%d] %T", i, a)
			if sp, ok := a.(*templparser.SpreadAttributes); ok {
				t.Logf("  spread expr value=%q range=%v", sp.Expression.Value, sp.Expression.Range)
			}
		}
	}
}
