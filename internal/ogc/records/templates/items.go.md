{{- /*gotype: github.com/PDOK/gokoala/internal/engine.TemplateData*/ -}}
# {{ .Params.CollectionID }} records

{{ .Params.NumberMatched }} matching records

{{range .Params.Items}}
- [{{.Title}}]({{.Href}}) (`{{.ID}}`){{if .Description}}: {{.Description}}{{end}}
{{else}}
No records found.
{{end}}