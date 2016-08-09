package templates

import (
	"fmt"
)

var builtinTemplates = map[string]func() string{
	"google": google_GenerateTemplate,
}

func GetBuiltinTemplates() []string {
	rval := []string{}

	for k, _ := range builtinTemplates {
		rval = append(rval, k)
	}

	return rval
}

func GetBuiltinTemplate(name string) (string, error) {
	if _, ok := builtinTemplates[name]; !ok {
		return "", fmt.Errorf("Invalid builtin template name: %s", name)
	}

	return builtinTemplates[name](), nil
}
