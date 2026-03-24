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

func writeCSV(filePath string, examples []example) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	w.WriteString("text,label,root,difficulty,transforms,category\n")
	for _, ex := range examples {
		escapedText := `"` + strings.ReplaceAll(ex.Text, `"`, `""`) + `"`
		transforms := strings.Join(ex.Transforms, ";")
		fmt.Fprintf(w, "%s,%d,%s,%s,%s,%s\n",
			escapedText, ex.Label, ex.Root, ex.Difficulty, transforms, ex.Category)
	}
	return w.Flush()
}
