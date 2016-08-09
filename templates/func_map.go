package templates

import (
	"text/template"
)

func DefaultFunctionMap() template.FuncMap {
	return template.FuncMap{
		"default":  nopFunc,
		"long":     nopFunc,
		"optional": nopFunc,
		"required": nopFunc,
		"short":    nopFunc,
		"variable": nopFunc,
	}
}

func nopFunc(args ...interface{}) interface{} {
	return nil
}
