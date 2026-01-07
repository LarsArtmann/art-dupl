package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/cli"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/job"
	"github.com/LarsArtmann/art-dupl/printer"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/util"
	"github.com/spf13/cobra"
)

func Run() int { //nolint:cyclop,funlen // Main CLI entry point with error handling paths
	flag.Usage = func() { fmt.Fprintln(os.Stderr, `Usage: art-dupl [flags] [paths]`) }
	cliCfg := cli.NewCLIConfig()

	flag.Parse()

	var fileConfig *config.Config
	var err error
	if *cliCfg.ConfigFile != "" {
		fileConfig, err = config.LoadConfig(*cliCfg.ConfigFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
			os.Exit(1)
			return 1
		}
	}
	appConfig := &config.Config{}

	if cliCfg.IsVerbose() {
		appConfig.Verbose = true
	}

	threshold := cliCfg.GetThreshold()
	if threshold != cli.DefaultThreshold {
		appConfig.Threshold = threshold
	}

	if *cliCfg.Vendor {
		appConfig.IncludeVendor = *cliCfg.Vendor
	}
	if *cliCfg.Files {
		appConfig.FilesFromStdin = *cliCfg.Files
	}

	switch {
	case *cliCfg.HTML:
		appConfig.OutputFormat = config.OutputFormatHTML
	case *cliCfg.Plumbing:
		appConfig.OutputFormat = config.OutputFormatPlumbing
	case *cliCfg.JSONFlag:
		appConfig.OutputFormat = config.OutputFormatJSON
	}

	if flag.NArg() > 0 {
		appConfig.Paths = flag.Args()
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if err = config.ValidateConfig(mergedConfig); err != nil {
		fmt.Fprintf(os.Stderr, "configuration error: %v\n", err)
		os.Exit(1)
		return 1
	}

	if *cliCfg.HTML && *cliCfg.Plumbing {
		fmt.Fprintf(os.Stderr, "error: you can have either plumbing or HTML output\n")
		os.Exit(1)
		return 1
	}
	if *cliCfg.HTML && *cliCfg.JSONFlag {
		fmt.Fprintf(os.Stderr, "error: you can have either HTML or JSON output\n")
		os.Exit(1)
		return 1
	}
	if *cliCfg.Plumbing && *cliCfg.JSONFlag {
		fmt.Fprintf(os.Stderr, "error: you can have either plumbing or JSON output\n")
		os.Exit(1)
		return 1
	}

	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		fmt.Fprintf(os.Stderr, "analysis error: %v\n", err)
		os.Exit(1)
		return 1
	}

	outputWriter := os.Stdout
	var outputFile *os.File
	if mergedConfig.OutputFile != "" {
		file, err := os.Create(mergedConfig.OutputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
			os.Exit(1)
			return 1
		}
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", closeErr)
			}
		}()
		outputFile = file
		outputWriter = file
	}

	p := createPrinter(mergedConfig.OutputFormat)(outputWriter, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, *cliCfg.SortBy, mergedConfig.Threshold); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		if outputFile != nil {
			_ = outputFile.Close()
		}
		os.Exit(1) //nolint:gocritic // Defer already executed before this point
		return 1
	}
	return 0
}

func createPrinter(outputFormat config.OutputFormat) func(io.Writer, printer.ReadFile) printer.Printer {
	switch outputFormat {
	case config.OutputFormatHTML:
		return printer.NewHTML
	case config.OutputFormatPlumbing:
		return printer.NewPlumbing
	case config.OutputFormatJSON:
		return printer.NewJSON
	case config.OutputFormatText:
		return printer.NewText
	default:
		return printer.NewText
	}
}

func buildSuffixTree(paths []string, verbose, filesFromStdin bool) (*suffixtree.STree, []*syntax.Node, int, error) {
	if verbose {
		log.Println("Building suffix tree")
	} else {
		fmt.Fprintf(os.Stderr, "    📖 Parsing files and building analysis tree...")
	}

	schan, filesCountChan := job.Parse(filesFeedWithOptions(paths, filesFromStdin))
	t, data, done := job.BuildTree(schan)
	<-done

	filesCount := <-filesCountChan
	t.Update(&syntax.Node{Type: -1})

	if verbose {
		log.Println("Searching for clones")
	} else {
		fmt.Fprintf(os.Stderr, " ✅\n")
	}

	return t, *data, filesCount, nil
}

func filesFeedWithOptions(paths []string, fromStdin bool) chan string {
	if fromStdin {
		fchan := make(chan string)
		go func() {
			s := bufio.NewScanner(os.Stdin)
			for s.Scan() {
				f := s.Text()
				fchan <- strings.TrimPrefix(f, "./")
			}
			close(fchan)
		}()
		return fchan
	}
	return crawlPaths(paths)
}

func crawlPaths(paths []string) chan string { //nolint:cyclop // Path crawling with multiple error handling paths
	fchan := make(chan string)
	go func() {
		for _, path := range paths {
			info, err := os.Lstat(path)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot stat %s: %v\n", path, err)
				os.Exit(1)
			}
			if !info.IsDir() {
				fchan <- path
				continue
			}
			err = filepath.Walk(path, func(path string, info os.FileInfo, _ error) error {
				// Check vendor flag using flag lookup to avoid redefinition
				vendorFlag := flag.Lookup("vendor")
				includeVendor := vendorFlag == nil || vendorFlag.Value.(flag.Getter).Get().(bool)
				if !includeVendor && (strings.HasPrefix(path, cli.VendorDirPrefix) ||
					strings.Contains(path, cli.VendorDirInPath)) {
					return nil
				}
				if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
					fchan <- path
				}
				return nil
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: cannot walk %s: %v\n", path, err)
				os.Exit(1)
			}
		}
		close(fchan)
	}()
	return fchan
}

func executeAnalysis(cfg *config.Config, paths []string) (chan syntax.Match, int, error) {
	t, data, filesCount, err := buildSuffixTree(paths, cfg.Verbose, cfg.FilesFromStdin)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build suffix tree for paths %v: %w", paths, err)
	}

	multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
	duplChan := make(chan syntax.Match)

	go func() {
		defer close(duplChan)
		matches := multiDetector.FindDuplOver(cfg.Threshold)
		for match := range matches {
			duplChan <- match
		}
	}()

	return duplChan, filesCount, nil
}

func printDupls(p printer.Printer, duplChan <-chan syntax.Match, sortBy string, threshold int) error { //nolint:cyclop // Output formatting with multiple conditional paths
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	if err := p.PrintHeader(); err != nil {
		return fmt.Errorf("failed to print header (sortBy: %s, threshold: %d): %w", sortBy, threshold, err)
	}

	for _, k := range keys {
		uniq := util.Unique(groups[k])
		if len(uniq) > 1 {
			if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
				jsonPrinter.SetHash(k)
			}
			if err := p.PrintClones(uniq, sortBy); err != nil {
				return fmt.Errorf("failed to print clones for hash %s (sortBy: %s): %w", k, sortBy, err)
			}
		}
	}

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		if err := jsonPrinter.OutputJSON(threshold, sortBy); err != nil {
			return fmt.Errorf("failed to output JSON (threshold: %d, sortBy: %s): %w", threshold, sortBy, err)
		}
	}

	if err := p.PrintFooter(); err != nil {
		return fmt.Errorf("failed to print footer: %w", err)
	}
	return nil
}

func runCobraCommand(cmd *cobra.Command, args []string) error { //nolint:cyclop,funlen // Cobra command with multiple flag handling paths
	configFile, _ := cmd.Flags().GetString("config")
	vendor, _ := cmd.Flags().GetBool("vendor")
	verbose, _ := cmd.Flags().GetBool("verbose")
	threshold, _ := cmd.Flags().GetInt("threshold")
	files, _ := cmd.Flags().GetBool("files")
	html, _ := cmd.Flags().GetBool("html")
	jsonFlag, _ := cmd.Flags().GetBool("json")
	plumbing, _ := cmd.Flags().GetBool("plumbing")
	sortBy, _ := cmd.Flags().GetString("sort")
	allFlag, _ := cmd.Flags().GetBool("all")
	_, _ = cmd.Flags().GetString("output-dir")
	profile, _ := cmd.Flags().GetBool("profile")
	timeoutStr, _ := cmd.Flags().GetString("timeout")

	var fileConfig *config.Config
	var err error
	if configFile != "" {
		fileConfig, err = config.LoadConfig(configFile)
		if err != nil {
			return fmt.Errorf("error loading config from file %q: %w", configFile, err)
		}
	}

	appConfig := &config.Config{}

	if verbose {
		appConfig.Verbose = true
	}

	if threshold != 15 {
		appConfig.Threshold = threshold
	}

	if vendor {
		appConfig.IncludeVendor = vendor
	}
	if files {
		appConfig.FilesFromStdin = files
	}

	switch {
	case html:
		appConfig.OutputFormat = config.OutputFormatHTML
	case plumbing:
		appConfig.OutputFormat = config.OutputFormatPlumbing
	case jsonFlag:
		appConfig.OutputFormat = config.OutputFormatJSON
	}

	if profile {
		appConfig.Profile = true
	}

	// Parse timeout duration (e.g., "30m", "1h", "2h30m")
	if timeoutStr != "30m" && timeoutStr != "" {
		duration, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout format %q (use '30m', '1h', etc.): %w", timeoutStr, err)
		}
		appConfig.Timeout = int(duration.Seconds())
	}

	if len(args) > 0 {
		appConfig.Paths = args
	}

	mergedConfig := config.MergeConfigs(fileConfig, appConfig)

	if err = config.ValidateConfig(mergedConfig); err != nil {
		return fmt.Errorf("configuration validation failed (paths: %v): %w", mergedConfig.Paths, err)
	}

	if allFlag {
		return errors.New("all mode not yet implemented")
	}

	duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
	if err != nil {
		return fmt.Errorf("analysis failed for paths %v: %w", mergedConfig.Paths, err)
	}

	p := createPrinter(mergedConfig.OutputFormat)(os.Stdout, os.ReadFile)

	if jsonPrinter, ok := p.(*printer.JSONPrinter); ok {
		jsonPrinter.SetFilesCount(filesCount)
	}

	if err := printDupls(p, duplChan, sortBy, mergedConfig.Threshold); err != nil {
		return fmt.Errorf("failed to print duplicates (sortBy: %s, threshold: %d): %w", sortBy, mergedConfig.Threshold, err)
	}

	return nil
}
