package templ

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// benchmarkTemplSrc is a representative component file: declarations,
// flow control, conditional attributes, expressions, CSS template, and raw
// Go code — every construct whose parsing changed with expression-aware
// detection (ADR-0027).
const benchmarkTemplSrc = `package components

import "fmt"

type User struct {
	Name  string
	Kind  string
	Admin bool
}

css panelStyle() {
	background-color: #fff;
	padding: 1rem;
}

templ header(user User) {
	<header class={ panelStyle() }>
		<h1>{ user.Name }</h1>
		if user.Admin {
			<span class="badge">admin</span>
		} else {
			<span>guest</span>
		}
	</header>
}

templ list(items []string) {
	<ul>
		for _, item := range items {
			<li>{ item }</li>
		}
	</ul>
}

templ panel(user User) {
	<div class={ panelStyle() }>
		switch user.Kind {
			case "admin":
				@header(user)
			case "guest":
				<p>restricted</p>
			default:
				<p>unknown</p>
		}
		if user.Kind != "" {
			@list([]string{"a", "b"})
		}
		{{ fmt.Println(user.Name) }}
		<input type="text" value={ user.Name } if user.Admin { disabled } else { readonly }/>
	</div>
}
`

func BenchmarkTemplParse(b *testing.B) {
	modes := []struct {
		name string
		mode golang.DetectionMode
	}{
		{"semantic", golang.DetectionModeSemantic},
		{"exact", golang.DetectionModeExact},
		{"structural", golang.DetectionModeStructural},
	}

	for _, bm := range modes {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()

			for b.Loop() {
				node, _, err := ParseBytesWithMode("bench.templ", []byte(benchmarkTemplSrc), bm.mode)
				if err != nil {
					b.Fatalf("ParseBytesWithMode failed: %v", err)
				}

				if node == nil {
					b.Fatal("ParseBytesWithMode returned nil")
				}

				_ = syntax.Serialize(node)
			}
		})
	}
}
