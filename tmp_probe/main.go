package main

import (
	"fmt"

	"github.com/LarsArtmann/gogenfilter/v3"
)

func main() {
	for _, p := range []string{
		"/tmp/art-dupl-bdd-x/pkg2/discard1.go",
		"C:/Users/RUNNER~1/AppData/Local/Temp/art-dupl-bdd-x/pkg2/discard1.go",
		"pkg2/discard1.go",
	} {
		fmt.Printf("path=%q match=%v\n", p, gogenfilter.MatchPattern(p, "pkg2/*"))
	}
}
