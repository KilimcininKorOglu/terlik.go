package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		switch {
		case arg == "--lang" && i+1 < len(args):
			i++
			opts.lang = args[i]
		case arg == "--pos" && i+1 < len(args):
			i++
			opts.pos, _ = strconv.Atoi(args[i])
		case arg == "--neg" && i+1 < len(args):
			i++
			opts.neg, _ = strconv.Atoi(args[i])
		case arg == "--seed" && i+1 < len(args):
			i++
			opts.seed, _ = strconv.Atoi(args[i])
		case arg == "--out" && i+1 < len(args):
			i++
			opts.out = args[i]
		case arg == "--format" && i+1 < len(args):
			i++
			opts.format = args[i]
		case arg == "--difficulty" && i+1 < len(args):
			i++
			opts.difficulty = args[i]
		case arg == "--stats":
			opts.stats = true
		case arg == "--validate":
			opts.validate = true
		case arg == "--dry-run":
			opts.dryRun = true
		default:
			// Language shortcuts: --tr, --en, --es, --de
			for _, lang := range supportedLangs {
				if arg == "--"+lang {
					opts.lang = lang
					break
				}
			}
		}
	}

	if opts.lang == "" {
		fmt.Fprintln(os.Stderr, "ERROR: Language must be specified. Usage: --lang tr or --tr")
		fmt.Fprintln(os.Stderr, "Supported languages: tr, en, es, de")
		os.Exit(1)
	}

	supported := false
	for _, l := range supportedLangs {
		if opts.lang == l {
			supported = true
			break
		}
	}
	if !supported {
		fmt.Fprintf(os.Stderr, "ERROR: Unsupported language: %s\n", opts.lang)
		fmt.Fprintln(os.Stderr, "Supported languages: tr, en, es, de")
		os.Exit(1)
	}

	if opts.format != "jsonl" && opts.format != "csv" && opts.format != "both" {
		fmt.Fprintf(os.Stderr, "ERROR: Invalid format: %s. Use jsonl, csv, or both.\n", opts.format)
		os.Exit(1)
	}

	return opts
}

func main() {
	opts := parseArgs(os.Args[1:])
	rand := mulberry32(opts.seed)
	lang := langConfigs[opts.lang]

	fmt.Println()
	fmt.Println("  Synthetic Profanity Dataset Generator")
	fmt.Printf("  Language: %s (%s)\n", lang.name, opts.lang)
	fmt.Printf("  Positive: %d, Negative: %d\n", opts.pos, opts.neg)
	fmt.Printf("  Seed: %d, Format: %s\n", opts.seed, opts.format)
	fmt.Printf("  Difficulty: %s\n", opts.difficulty)

	// Load data
	fmt.Println("\n  Loading data files...")
	data, err := loadAllData(opts.lang)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}

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

	if opts.dryRun {
		fmt.Println("\n  [DRY-RUN] Data files loaded successfully. No generation performed.")
		if opts.stats {
			var sampleExamples []example
			for i := 0; i < 50; i++ {
				sampleExamples = append(sampleExamples, renderPositiveExample(data, lang, rand))
			}
			for i := 0; i < 50; i++ {
				sampleExamples = append(sampleExamples, renderNegativeExample(data, rand))
			}
			printStats(sampleExamples, opts.lang)
			printSamples(sampleExamples, rand)
		}
		return
	}

	// Generate examples with deduplication
	fmt.Println("\n  Generating examples...")
	startTime := time.Now()
	var examples []example
	seen := make(map[string]bool)
	maxRetries := 50

	for i := 0; i < opts.pos; i++ {
		var ex example
		retries := maxRetries
		for retries > 0 {
			ex = renderPositiveExample(data, lang, rand)
			if opts.difficulty != "all" && ex.Difficulty != opts.difficulty {
				retries--
				continue
			}
			if !seen[ex.Text] {
				break
			}
			retries--
		}
		seen[ex.Text] = true
		examples = append(examples, ex)
	}

	for i := 0; i < opts.neg; i++ {
		var ex example
		retries := maxRetries
		for retries > 0 {
			ex = renderNegativeExample(data, rand)
			if !seen[ex.Text] {
				break
			}
			retries--
		}
		seen[ex.Text] = true
		examples = append(examples, ex)
	}

	genTime := time.Since(startTime)
	fmt.Printf("  %d examples generated (%s)\n", len(examples), genTime.Round(time.Millisecond))

	// Shuffle
	shuffle(examples, rand)

	// Write output
	if err := os.MkdirAll(opts.out, 0o755); err != nil {
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

	// Stats
	if opts.stats {
		printStats(examples, opts.lang)
	}

	// Validation
	if opts.validate {
		errs := validateExamples(examples)
		if errs > 0 {
			fmt.Printf("\n  [WARN] %d errors found. Review the output.\n", errs)
		} else {
			fmt.Println("\n  [OK] All validation checks passed.")
		}
	}

	// Samples
	printSamples(examples, rand)

	fmt.Println("\n  Done.")
}
