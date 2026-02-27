package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
)

// ANSI colour codes (disabled automatically when output is not a terminal).
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiCyan   = "\033[36m"
)

// isTerminal returns true when stdout is a real terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func colored(code, s string) string {
	if !isTerminal() {
		return s
	}
	return code + s + ansiReset
}

// Bold wraps s in ANSI bold (no-op when not a terminal).
func Bold(s string) string { return colored(ansiBold, s) }

// Green wraps s in ANSI green.
func Green(s string) string { return colored(ansiGreen, s) }

// Yellow wraps s in ANSI yellow.
func Yellow(s string) string { return colored(ansiYellow, s) }

// Red wraps s in ANSI red.
func Red(s string) string { return colored(ansiRed, s) }

// Cyan wraps s in ANSI cyan.
func Cyan(s string) string { return colored(ansiCyan, s) }

// PrintTable renders rows in a nicely aligned tabwriter table.
// headers are printed in bold; use wideMode=true to include Wide columns.
func PrintTable(res *ResourceDef, items []map[string]interface{}, wideMode bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	var cols []Column
	for _, c := range res.Columns {
		if !c.Wide || wideMode {
			cols = append(cols, c)
		}
	}

	// Header row
	headers := make([]string, len(cols))
	for i, c := range cols {
		headers[i] = Bold(c.Header)
	}
	fmt.Fprintln(w, strings.Join(headers, "\t"))

	// Data rows
	for _, item := range items {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = fieldValue(item, c.Key)
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	w.Flush()
}

// PrintJSON pretty-prints any value as indented JSON to stdout.
func PrintJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// PrintDescribe prints a resource in a kubectl-describe-style key:value layout.
func PrintDescribe(obj map[string]interface{}) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Priority fields shown first
	priority := []string{
		"displayName", "id", "@odata.type", "description",
		"createdDateTime", "lastModifiedDateTime",
	}
	printed := map[string]bool{}

	for _, k := range priority {
		if v, ok := obj[k]; ok {
			printed[k] = true
			printKV(w, k, v)
		}
	}

	// Remaining fields in alphabetical order
	keys := sortedKeys(obj)
	for _, k := range keys {
		if printed[k] {
			continue
		}
		printKV(w, k, obj[k])
	}
	w.Flush()
}

func printKV(w *tabwriter.Writer, key string, val interface{}) {
	var display string
	switch v := val.(type) {
	case string:
		display = v
	case bool:
		if v {
			display = Green("true")
		} else {
			display = Yellow("false")
		}
	case float64:
		display = fmt.Sprintf("%.0f", v)
	case nil:
		display = "<null>"
	case []interface{}:
		if len(v) == 0 {
			display = "[]"
		} else {
			parts := make([]string, 0, len(v))
			for _, item := range v {
				parts = append(parts, fmt.Sprintf("%v", item))
			}
			display = strings.Join(parts, ", ")
		}
	default:
		b, _ := json.Marshal(v)
		display = string(b)
	}
	fmt.Fprintf(w, "%s\t%s\n", Bold(key+":"), display)
}

// PrintSuccess prints a success message to stdout.
func PrintSuccess(format string, args ...interface{}) {
	fmt.Printf(Green("✓ ")+format+"\n", args...)
}

// PrintError prints an error to stderr.
func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, Red("✗ ")+format+"\n", args...)
}

// PrintWarning prints a warning to stderr.
func PrintWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, Yellow("⚠ ")+format+"\n", args...)
}

// sortedKeys returns map keys sorted alphabetically.
func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// simple insertion sort (maps are small)
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
