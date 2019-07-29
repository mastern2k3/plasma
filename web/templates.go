package web

import (
	"text/template"
)

type HomeObjectModel struct {
	Path  string
	Error error
}

type HomeModel struct {
	Objects []HomeObjectModel
}

var HomeTemplate = template.Must(template.New("home").Parse(`
<html><body><pre>
Welcome to Plasma
=================

  objects:

{{range .Objects}}    <a href="{{.Path}}">{{.Path}}</a>{{if .Error}} [✘ with error]{{end}}
{{end}}
</pre></body></html>
`))
