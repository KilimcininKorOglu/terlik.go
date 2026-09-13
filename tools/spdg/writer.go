package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func writeJSONL(filePath string, examples []example) error {
	// Path comes from the operator's --out CLI flag, not untrusted input.
	f, err := os.Create(filePath) // #nosec G304
	if err != nil {
		return err
	}
	// Write errors are sticky in bufio.Writer and returned by w.Flush() below.
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	for _, ex := range examples {
		data, err := json.Marshal(ex)
		if err != nil {
			return err
		}
		_, _ = w.Write(data)  // error is sticky; surfaced by w.Flush()
		_ = w.WriteByte('\n') // error is sticky; surfaced by w.Flush()
	}
	return w.Flush()
}

func csvQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func writeCSV(filePath string, examples []example) error {
	// Path comes from the operator's --out CLI flag, not untrusted input.
	f, err := os.Create(filePath) // #nosec G304
	if err != nil {
		return err
	}
	// Write errors are sticky in bufio.Writer and returned by w.Flush() below.
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)
	_, _ = w.WriteString("text,label,root,difficulty,transforms,category\n") // error is sticky; surfaced by w.Flush()
	for _, ex := range examples {
		transforms := strings.Join(ex.Transforms, ";")
		_, _ = fmt.Fprintf(w, "%s,%d,%s,%s,%s,%s\n",
			csvQuote(ex.Text), ex.Label, csvQuote(ex.Root),
			csvQuote(ex.Difficulty), csvQuote(transforms), csvQuote(ex.Category)) // error is sticky; surfaced by w.Flush()
	}
	return w.Flush()
}
