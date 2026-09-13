package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

type cliOptions struct {
	lang       string
	pos        int
	neg        int
	seed       int
	out        string
	format     string
	stats      bool
	difficulty string
	validate   bool
	dryRun     bool
}

var supportedLangs = []string{"tr", "en", "es", "de"}

func parseArgs(args []string) *cliOptions {
	opts := &cliOptions{
		pos:        20000,
		neg:        20000,
		seed:       42,
		out:        "output",
		format:     "jsonl",
		difficulty: "all",
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		next := ""
		hasNext := i+1 < len(args)
		if hasNext {
			next = args[i+1]
		}
		if applyValueFlag(opts, arg, next, hasNext) {
			i++
		} else if !applyBoolFlag(opts, arg) {
			applyLangShortcut(opts, arg)
		}
	}

	validateOptions(opts)
	return opts
}

// applyValueFlag handles "--flag value" arguments and reports whether it matched.
// A value flag at the end of the argument list (hasValue=false) falls through to
// the other handlers, exactly like the original switch did.
func applyValueFlag(opts *cliOptions, arg, next string, hasValue bool) bool {
	if !hasValue {
		return false
	}
	switch arg {
	case "--lang":
		opts.lang = next
	case "--pos":
		opts.pos = requireInt(next, "--pos")
	case "--neg":
		opts.neg = requireInt(next, "--neg")
	case "--seed":
		opts.seed = requireInt(next, "--seed")
	case "--out":
		opts.out = next
	case "--format":
		opts.format = next
	case "--difficulty":
		opts.difficulty = next
	default:
		return false
	}
	return true
}

// applyBoolFlag handles the no-value flags and reports whether it matched.
func applyBoolFlag(opts *cliOptions, arg string) bool {
	switch arg {
	case "--stats":
		opts.stats = true
	case "--validate":
		opts.validate = true
	case "--dry-run":
		opts.dryRun = true
	default:
		return false
	}
	return true
}

// applyLangShortcut maps "--tr"-style shortcuts onto opts.lang; unknown args stay ignored.
func applyLangShortcut(opts *cliOptions, arg string) {
	for _, lang := range supportedLangs {
		if arg == "--"+lang {
			opts.lang = lang
			return
		}
	}
}

// requireInt parses a numeric flag value, exiting with the CLI's usage error on bad input.
func requireInt(val, flag string) int {
	n, err := strconv.Atoi(val)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid %s value: %s\n", flag, val)
		os.Exit(1)
	}
	return n
}

// validateOptions enforces the language and format constraints after parsing.
func validateOptions(opts *cliOptions) {
	if opts.lang == "" {
		fmt.Fprintln(os.Stderr, "ERROR: Language must be specified. Usage: --lang tr or --tr")
		fmt.Fprintln(os.Stderr, "Supported languages: tr, en, es, de")
		os.Exit(1)
	}

	if !isSupportedLang(opts.lang) {
		fmt.Fprintf(os.Stderr, "ERROR: Unsupported language: %s\n", opts.lang)
		fmt.Fprintln(os.Stderr, "Supported languages: tr, en, es, de")
		os.Exit(1)
	}

	if opts.format != "jsonl" && opts.format != "csv" && opts.format != "both" {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid format: %s. Use jsonl, csv, or both.\n", opts.format)
		os.Exit(1)
	}
}

func isSupportedLang(lang string) bool {
	return slices.Contains(supportedLangs, lang)
}

func main() {
	opts := parseArgs(os.Args[1:])
	rand := mulberry32(opts.seed)
	lang := langConfigs[opts.lang]

	printBanner(opts, lang)

	fmt.Println("\n  Loading data files...")
	data, err := loadAllData(opts.lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	printDataSummary(data)

	if opts.dryRun {
		runDryRun(opts, data, lang, rand)
		return
	}

	examples := generateExamples(opts, data, lang, rand)
	shuffle(examples, rand)
	writeOutputs(opts, examples)
	finishReport(opts, examples, rand)
}

// printBanner writes the startup banner exactly as the original main did.
func printBanner(opts *cliOptions, lang *langConfig) {
	fmt.Println()
	fmt.Println("  Synthetic Profanity Dataset Generator")
	fmt.Printf("  Language: %s (%s)\n", lang.name, opts.lang)
	fmt.Printf("  Positive: %d, Negative: %d\n", opts.pos, opts.neg)
	fmt.Printf("  Seed: %d, Format: %s\n", opts.seed, opts.format)
	fmt.Printf("  Difficulty: %s\n", opts.difficulty)
}

// printDataSummary reports the loaded dataset sizes.
func printDataSummary(data *dataSet) {
	fmt.Println("  Loading complete:")
	fmt.Printf("    Positive roots:     %d\n", len(data.rootsPositive))
	fmt.Printf("    Negative roots:     %d\n", len(data.rootsNegative))
	fmt.Printf("    Template (pos):     %d\n", len(data.templatesPositive))
	fmt.Printf("    Template (neg):     %d\n", len(data.templatesNegative))
	fmt.Printf("    Context (pos):      %d\n", len(data.contextsPositive))
	fmt.Printf("    Context (neg):      %d\n", len(data.contextsNegative))
	fmt.Printf("    Suffix:             %d\n", len(data.suffixes))
	fmt.Printf("    Leet map:           %d chars\n", len(data.leetMap))
	fmt.Printf("    Emoji:              %d\n", len(data.emojiReplacements))
	fmt.Printf("    Separators:         %d\n", len(data.separators))
	fmt.Printf("    Unicode map:        %d chars\n", len(data.unicodeMap))
	fmt.Printf("    Zalgo chars:        %d\n", len(data.zalgoChars))
	fmt.Printf("    ZWC chars:          %d\n", len(data.zwcChars))
}

// runDryRun renders optional stats from a small sample without writing output.
func runDryRun(opts *cliOptions, data *dataSet, lang *langConfig, rand func() float64) {
	fmt.Println("\n  [DRY-RUN] Data files loaded successfully. No generation performed.")
	if opts.stats {
		var sampleExamples []example
		for range 50 {
			sampleExamples = append(sampleExamples, renderPositiveExample(data, lang, rand))
		}
		for range 50 {
			sampleExamples = append(sampleExamples, renderNegativeExample(data, rand))
		}
		printStats(sampleExamples, opts.lang)
		printSamples(sampleExamples, rand)
	}
}

// generateExamples renders the requested positive and negative examples with
// cross-set deduplication, preserving the original PRNG call order.
func generateExamples(opts *cliOptions, data *dataSet, lang *langConfig, rand func() float64) []example {
	fmt.Println("\n  Generating examples...")
	startTime := time.Now()
	seen := make(map[string]bool)
	examples := generatePositives(opts, data, lang, rand, seen)
	examples = append(examples, generateNegatives(opts, data, rand, seen)...)
	genTime := time.Since(startTime)
	fmt.Printf("  %d examples generated (%s)\n", len(examples), genTime.Round(time.Millisecond))
	return examples
}

// generatePositives renders opts.pos positive examples, re-rendering entries that
// miss the requested difficulty or duplicate an already generated text.
func generatePositives(opts *cliOptions, data *dataSet, lang *langConfig, rand func() float64, seen map[string]bool) []example {
	maxRetries := 50
	var examples []example
	for i := 0; i < opts.pos; i++ {
		var ex example
		found := false
		retries := maxRetries
		for retries > 0 {
			ex = renderPositiveExample(data, lang, rand)
			if opts.difficulty != "all" && ex.Difficulty != opts.difficulty {
				retries--
				continue
			}
			if !seen[ex.Text] {
				found = true
				break
			}
			retries--
		}
		if found {
			seen[ex.Text] = true
			examples = append(examples, ex)
		}
	}
	return examples
}

// generateNegatives renders opts.neg negative examples with dedup retries.
func generateNegatives(opts *cliOptions, data *dataSet, rand func() float64, seen map[string]bool) []example {
	maxRetries := 50
	var examples []example
	for i := 0; i < opts.neg; i++ {
		var ex example
		found := false
		retries := maxRetries
		for retries > 0 {
			ex = renderNegativeExample(data, rand)
			if !seen[ex.Text] {
				found = true
				break
			}
			retries--
		}
		if found {
			seen[ex.Text] = true
			examples = append(examples, ex)
		}
	}
	return examples
}

// writeOutputs creates the output directory and writes the selected formats.
func writeOutputs(opts *cliOptions, examples []example) {
	if err := os.MkdirAll(opts.out, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot create output directory: %v\n", err)
		os.Exit(1)
	}

	if opts.format == "jsonl" || opts.format == "both" {
		jsonlPath := filepath.Join(opts.out, fmt.Sprintf("export-%s.jsonl", opts.lang))
		fmt.Printf("  Writing: %s\n", jsonlPath)
		if err := writeJSONL(jsonlPath, examples); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to write JSONL: %v\n", err)
			os.Exit(1)
		}
	}

	if opts.format == "csv" || opts.format == "both" {
		csvPath := filepath.Join(opts.out, fmt.Sprintf("export-%s.csv", opts.lang))
		fmt.Printf("  Writing: %s\n", csvPath)
		if err := writeCSV(csvPath, examples); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: Failed to write CSV: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Println("  Write complete.")
}

// finishReport prints post-write stats, validation results, and samples.
func finishReport(opts *cliOptions, examples []example, rand func() float64) {
	if opts.stats {
		printStats(examples, opts.lang)
	}

	if opts.validate {
		errs := validateExamples(examples)
		if errs > 0 {
			fmt.Printf("\n  [WARN] %d errors found. Review the output.\n", errs)
		} else {
			fmt.Println("\n  [OK] All validation checks passed.")
		}
	}

	printSamples(examples, rand)

	fmt.Println("\n  Done.")
}
