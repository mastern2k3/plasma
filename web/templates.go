package web

import (
	"text/template"
)

type HomeObjectModel struct {
	Path string
}

type HomeModel struct {
	Objects []HomeObjectModel
}

var HomeTemplate = template.Must(template.New("home").Parse(`
<html><body><pre>
Welcome to Plasma
=================

  objects:

{{range .Objects}}    <a href="{{.Path}}">{{.Path}}</a>
{{end}}
</pre></body></html>
`))
