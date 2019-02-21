package bel

import (
	"bytes"
	"fmt"
	"io"
	"text/template"
)

const interfaceTemplate = `
{{ define "iface" -}}
{
    {{ range .Members -}}
    {{ .Name }}{{ if .IsOptional }}?{{ end }}: {{ subt .Type }}
    {{ end }}
}
{{ end -}}
{{- define "simple" }}{{ .Name }}{{ end -}}
{{- define "map" }}{ [key: {{ subt (mapKeyType .) }}]: foo }{{ end -}}
{{- define "array" }}{{ subt (arrType .) }}[]{{ end -}}
{{- define "root-enum" }}export enum {{ .Name }} {
    {{ range .EnumMembers }}{{ .Name }} = {{ .Value }}
    {{ end }}
}{{ end -}}
{{- define "root-iface" }} export interface {{ .Name }} {{ template "iface" . }} {{ end -}}
{{ subtroot . }}
`

func Render(types []TypescriptType, w io.Writer) error {
	for _, t := range types {
		err := t.Render(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ts *TypescriptType) Render(w io.Writer) error {
	getParam := func(nme string, idx, minlen int) func(t TypescriptType) (*TypescriptType, error) {
		return func(t TypescriptType) (*TypescriptType, error) {
			if len(t.Params) != minlen {
				return nil, fmt.Errorf("map needs %d type params", minlen)
			}
			return &t.Params[idx], nil
		}
	}

	var tpl *template.Template
	funcs := template.FuncMap{
		"mapKeyType": getParam("map", 0, 2),
		"mapValType": getParam("map", 1, 2),
		"arrType":    getParam("array", 0, 1),
		"subt": func(t TypescriptType) (string, error) {
			var b bytes.Buffer
			if err := tpl.ExecuteTemplate(&b, string(t.Kind), t); err != nil {
				return "", err
			}
			return b.String(), nil
		},
        "subtroot": func(t TypescriptType) (string, error) {
			var b bytes.Buffer
			if err := tpl.ExecuteTemplate(&b, "root-" + string(t.Kind), t); err != nil {
				return "", err
			}
			return b.String(), nil
		},
	}

	var err error
	tpl, err = template.New("interface").Funcs(funcs).Parse(interfaceTemplate)
	if err != nil {
		return err
	}

	return tpl.Execute(w, ts)
}
