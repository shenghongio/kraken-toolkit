package printer

import (
	"encoding/json"
	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v3"
	"os"
)

// Table represents generic tabular data
// Printer does not know anything about Redis,MySQL,Kafka....
// Business/commandlayer converts its result into Table.

type Table struct {
	Headers []string
	Rows    [][]string
}

// PrintTable renders a table to stdout

func PrintTable(data Table) error {
	writer := table.NewWriter()
	writer.SetOutputMirror(os.Stdout)

	// no heavy box borders
	writer.SetStyle(table.StyleLight)

	if len(data.Headers) > 0 {
		writer.AppendHeader(table.Row{data.Headers})
	}

	// Rows
	for _, row := range data.Rows {
		writer.AppendRow(table.Row{row})
	}
	writer.Render()
	return nil
}

func PrintJSON(v any) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func PrintYAML(v any) error {
	encoder := yaml.NewEncoder(os.Stdout)
	defer encoder.Close()
	return encoder.Encode(v)
}
