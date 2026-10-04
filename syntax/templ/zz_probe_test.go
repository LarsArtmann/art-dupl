package templ

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestProbeFixture(t *testing.T) {
	pieces := []struct{ name, src string }{
		{"cases-tab", "templ p(user User) {\n\t<div>\n\t\tswitch user.Kind {\n\t\t\tcase \"admin\":\n\t\t\t\t<p>a</p>\n\t\t\tdefault:\n\t\t\t\t<p>b</p>\n\t\t}\n\t</div>\n}\n"},
		{"cases-flush", "templ p(user User) {\n\t<div>\n\t\tswitch user.Kind {\n\t\tcase \"admin\":\n\t\t\t<p>a</p>\n\t\tdefault:\n\t\t\t<p>b</p>\n\t\t}\n\t</div>\n}\n"},
	}
	for _, pc := range pieces {
		_, _, err := ParseBytesWithMode("probe.templ", []byte(pc.src), golang.DetectionModeSemantic)
		t.Logf("%s: err=%v", pc.name, err)
	}
}
