{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
{{with .Params}}
# {{.Title}}

Record ID: `{{.ID}}`

{{if .Description}}{{.Description}}{{end}}

## Properties
{{range .Properties}}
- **{{.Name}}**: {{.Value}}
{{end}}

## Assets
{{range .Assets}}
- [{{if .Title}}{{.Title}}{{else}}{{.Key}}{{end}}]({{.Href}}){{if .MediaType}} ({{.MediaType}}){{end}}
{{else}}
No assets listed.
{{end}}
{{end}}