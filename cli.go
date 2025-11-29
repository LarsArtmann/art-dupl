package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/golangci/dupl/job"
	"github.com/golangci/dupl/printer"
	"github.com/golangci/dupl/syntax"
)

// CLIInterface defines the interface for CLI operations
type CLIInterface interface {
	Exit(code int)
	Stderr() io.Writer
	Stdout() io.Writer
}

// RealCLI is the production implementation of CLIInterface
type RealCLI struct{}

func (r *RealCLI) Exit(code int) { os.Exit(code) }
func (r *RealCLI) Stderr() io.Writer { return os.Stderr }
func (r *RealCLI) Stdout() io.Writer { return os.Stdout }

// TestCLI is the test implementation of CLIInterface
type TestCLI struct {
	ExitCode int
	StderrBuf strings.Builder
	StdoutBuf strings.Builder
}

func (t *TestCLI) Exit(code int) { t.ExitCode = code }
func (t *TestCLI) Stderr() io.Writer { return &t.StderrBuf }
func (t *TestCLI) Stdout() io.Writer { return &t.StdoutBuf }

var cli CLIInterface = &RealCLI{}

// Run is the main application logic
func Run() int {
	flag.Usage = usage
	flag.Parse()
	if *html && *plumbing {
		fmt.Fprintf(cli.Stderr(), "error: you can have either plumbing or HTML output\n")
		cli.Exit(1)
		return 1
	}
	if flag.NArg() > 0 {
		paths = flag.Args()
	}

	if *verbose {
		log.Println("Building suffix tree")
	}
	schan := job.Parse(filesFeed())
	t, data, done := job.BuildTree(schan)
	<-done

	// finish stream
	t.Update(&syntax.Node{Type: -1})

	if *verbose {
		log.Println("Searching for clones")
	}
	mchan := t.FindDuplOver(*threshold)
	duplChan := make(chan syntax.Match)
	go func() {
		for m := range mchan {
			match := syntax.FindSyntaxUnits(*data, m, *threshold)
			if len(match.Frags) > 0 {
				duplChan <- match
			}
		}
		close(duplChan)
	}()

	newPrinter := printer.NewText
	if *html {
		newPrinter = printer.NewHTML
	} else if *plumbing {
		newPrinter = printer.NewPlumbing
	}
	p := newPrinter(cli.Stdout(), os.ReadFile)
	if err := printDupls(p, duplChan); err != nil {
		fmt.Fprintf(cli.Stderr(), "error: %v\n", err)
		cli.Exit(1)
		return 1
	}
	return 0
}