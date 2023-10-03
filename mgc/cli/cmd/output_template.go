package cmd

import (
	"fmt"
	"os"

	"text/template"
)

type templateOutputFormatter struct{}

func (*templateOutputFormatter) Format(value any, options string) error {
	text := options
	tmpl, err := template.New("template").Parse(text)
	if err != nil {
		return fmt.Errorf("template output formatter: %w", err)
	}

	return tmpl.Execute(os.Stdout, value)
}

func (*templateOutputFormatter) Description() string {
	return `Format using https://pkg.go.dev/text/template. Use "template=your-template-here."` +
		` For more complex specifications, see "template-file".`
}

func init() {
	outputFormatters["template"] = &templateOutputFormatter{}
}
