package templates

import (
	"bytes"
	"text/template"
)

type Template struct {
	Name string
	Text string

	t    *template.Template
	vars *Variables
}

func New(name, tmpl string) *Template {
	rval := new(Template)

	rval.Name = name
	rval.Text = tmpl

	return rval
}

func (t *Template) Process(funcs template.FuncMap) (*bytes.Buffer, error) {

	rval := new(bytes.Buffer)
	tmpl, err := template.New(t.Name).Funcs(funcs).Parse(t.Text)
	if err != nil {
		return nil, err
	}
	t.t = tmpl

	if err := t.t.Execute(rval, nil); err != nil {
		return nil, err
	}

	return rval, nil
}
