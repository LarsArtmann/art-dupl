//go:build tools

// Package tools pins the templ generator inside the module graph: go mod
// tidy considers tag-guarded files, so the generator's dependencies land in
// the nix proxy cache and flake.nix preBuild can run
// `go run github.com/a-h/templ/cmd/templ generate` offline at exactly the
// go.mod templ version. The nixpkgs templ CLI floats ahead of go.mod and
// emits code newer runtimes cannot compile. The build tag keeps go build,
// vet and test from loading a main package as an import.
package tools

import (
	_ "github.com/a-h/templ/cmd/templ"
)
