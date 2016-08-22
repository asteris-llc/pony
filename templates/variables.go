package templates

import (
	"bytes"
	"fmt"
	"text/template"
)

type Variable struct {
	Name     string
	Short    string
	Long     string
	Default  string
	Optional bool
}
type Variables map[string]*Variable

func (t *Template) ReadVariables() error {
	t.vars = &Variables{}

	if _, err := t.Process(t.variableFunctions()); err != nil {
		return err
	}

	return nil
}

func (t *Template) GetVariables() map[string]*Variable {
	return *t.vars
}

func (t *Template) validateVarName(fname, name string) error {
	if name == "" {
		return fmt.Errorf("No `variable` directive preceding `%s` directive in pipeline", fname)
	}

	// This shouldn't happen
	if _, ok := (*t.vars)[name]; !ok {
		return fmt.Errorf("Invalid variable name `%s` in `%s` directive in pipeline", name, fname)
	}

	return nil
}

func (t *Template) variableFunctions() template.FuncMap {
	rval := DefaultFunctionMap()

	rval["default"] = t.varDefaultFunc()
	rval["long"] = t.varLongFunc()
	rval["optional"] = t.varOptionalFunc()
	rval["required"] = t.varRequiredFunc()
	rval["short"] = t.varShortFunc()
	rval["variable"] = t.varVariableFunc()

	return rval
}

// varDefaultFunc()
// Assign a default value to a variable
//
// {{ variable "example" | default "foo" }}
//
func (t *Template) varDefaultFunc() func(def, varname string) (interface{}, error) {
	return func(def, varname string) (interface{}, error) {
		if err := t.validateVarName("default", varname); err != nil {
			return nil, err
		}

		(*t.vars)[varname].Default = def

		return varname, nil
	}
}

// varLongFunc()
// The `long` function takes the name of a {{ template }} block as its only argument.
// It processes the that template block and saves it to the Long member of the variable
// This is to allow a longer description of the Terraform variable
//
// {{- define "example_block" }}
// This is an example
// of a multi-line description
// {{ end -}}
// {{- variable "example" | long "example_block" -}}
//
func (t *Template) varLongFunc() func(block string, fname string) (interface{}, error) {
	return func(block string, varname string) (interface{}, error) {
		if err := t.validateVarName("long", varname); err != nil {
			return nil, err
		}

		if block == "" {
			return nil, fmt.Errorf("No template block name passed to `long` function")
		}

		b := new(bytes.Buffer)
		if err := t.t.ExecuteTemplate(b, block, nil); err != nil {
			return nil, err
		}

		(*t.vars)[varname].Long = b.String()

		return varname, nil
	}
}

func (t *Template) varOptionalFunc() func(varname string) (interface{}, error) {
	return func(varname string) (interface{}, error) {
		if err := t.validateVarName("optional", varname); err != nil {
			return nil, err
		}

		(*t.vars)[varname].Optional = true

		return varname, nil
	}
}

func (t *Template) varRequiredFunc() func(varname string) (interface{}, error) {
	return func(varname string) (interface{}, error) {
		if err := t.validateVarName("optional", varname); err != nil {
			return nil, err
		}

		(*t.vars)[varname].Optional = false

		return varname, nil
	}
}

// varShortFunc()
// Assign a short descriptive string to the variable
//
// {{ variable "example" | short "An example variable" }}
//
func (t *Template) varShortFunc() func(sdesc string, fname string) (interface{}, error) {
	return func(sdesc string, varname string) (interface{}, error) {
		if err := t.validateVarName("short", varname); err != nil {
			return nil, err
		}

		if sdesc == "" {
			return nil, fmt.Errorf("No description string block name passed to `short` function")
		}

		(*t.vars)[varname].Short = sdesc

		return varname, nil
	}
}

// varVariableFunc
// Create a Variable structure indexed by the passed in string.
// The string is returned to the pipeline so any following functions will
// be able to update the correct variable
//
func (t *Template) varVariableFunc() func(s string) (interface{}, error) {
	return func(s string) (interface{}, error) {
		if s == "" {
			return nil, fmt.Errorf("No variable name in `variable` function")
		}

		if _, ok := (*t.vars)[s]; ok {
			return nil, fmt.Errorf("Multiple uses of `variable %s`", s)
		}

		(*t.vars)[s] = &Variable{
			Name:     s,
			Optional: false,
		}

		return s, nil
	}
}
