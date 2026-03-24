package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func writeJSONL(filePath string, examples []example) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for _, ex := range examples {
		data, err := json.Marshal(ex)
		if err != nil {
			return err
		}
		w.Write(data)
		w.WriteByte('\n')
	}
	return w.Flush()
}

func csvQuote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}

func writeCSV(filePath string, examples []example) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("text,label,root,difficulty,transforms,category\n")
	for _, ex := range examples {
		transforms := strings.Join(ex.Transforms, ";")
		fmt.Fprintf(w, "%s,%d,%s,%s,%s,%s\n",
			csvQuote(ex.Text), ex.Label, csvQuote(ex.Root),
			csvQuote(ex.Difficulty), csvQuote(transforms), csvQuote(ex.Category))
	}
	return w.Flush()
}
