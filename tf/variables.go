package tf

import (
	"bytes"
	"fmt"
	"os"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/module"
	log "github.com/sirupsen/logrus"
)

const (
	_ = iota
	MetaRequired = "meta_required_variables"
)

var metaHandlers = []func(*Tf, *variables) error{
	metaRequiredHandler,
}

type variables struct {
	variables	map[string]*config.Variable
}

func (t *Tf) ReadVariables() error {
	vs := &variables{
		variables: make(map[string]*config.Variable),
	}

	vs.readVars(t.m)
	for k, v := range vs.variables {
		log.Debugf("var: %s, val: %+v", k, v)
	}

	os.Exit(0)

//	vs.Init(t)

	// Run through all of the meta variable handlers
	for _, mh := range metaHandlers {
		if err := mh(t, vs); err != nil {
			return err
		}
	}

	// The remaining variables are optional. Run through them here
	//
	if err := optionalVariables(t, vs); err != nil {
		return err
	}

	for k, v := range t.m.Config().Variables {
		fmt.Printf("k: %d, v: %+v\n",k, v)
	}

	return nil
}

func (vs *variables) Init(t *Tf) {
	vs.variables = make(map[string]*config.Variable, len(t.m.Config().Variables))

	for _, v := range t.m.Config().Variables {
		vs.variables[v.Name] = v
	}

}

func optionalVariables(t *Tf, vs *variables) error {
	if len(vs.variables) == 0 {
		log.Debug("No optional variables in module")

		// No more variables left
		return nil
	}

	if !t.cli.AskYesNo("Configure optional parameters? (y/N)", "n") {
		return nil
	}

	for _, v := range vs.variables {
		if err := askForValue(t, vs, v); err != nil {
			return err
		}
	}

	return nil
}

func metaRequiredHandler(t *Tf, vs *variables) error {
	if _, ok := vs.variables[MetaRequired]; !ok {
		// No required variables
		return nil
	}

	requiredListVar := vs.variables[MetaRequired]

	// Remove meta variable from variable map
	delete(vs.variables, MetaRequired)

	// Verify that the variable type is list
	if !isVariableType(requiredListVar, config.VariableTypeList) {
		return fmt.Errorf("Invalid type for meta_required_variables. \"%s\" != \"%s\"", 
					config.VariableTypeList, 
					requiredListVar.Type().Printable())
	}

	requiredList := requiredListVar.Default.([]interface{})
	for _, v := range requiredList {
		// variable names should only be strings
		switch v.(type) {
		case string:
		default:
			return fmt.Errorf("Invalid type for required variable name: '%T' != 'string'", v)
		}

		vname := v.(string)
		if _, ok := vs.variables[vname]; !ok {
			log.Warnf("Required variable '%s' not in module", vname)
			continue
		}

		requiredVar := vs.variables[vname]
		if err := askForValue(t, vs, requiredVar); err != nil {
			return err
		}
	}

	return nil
}

func isVariableType(v *config.Variable, t config.VariableType) bool {
	return v.Type() == t
}

func getDefault(v *config.Variable) string {
	var def string
	switch v.Default.(type) {
	case string:
		def = v.Default.(string)
	case nil:
		def = ""
	default:
		log.Warnf("Invalid type for %s default: %T", v.Name, v.Default)
		def = ""
	}

	return def
}

func buildPrompt(v *config.Variable, def string) string {
	prompt := bytes.NewBufferString("Enter a value for ")
	prompt.WriteString(v.Name)

	return prompt.String()
}

func askForValue(t *Tf, vs *variables, v *config.Variable) error {
	if !isVariableType(v, config.VariableTypeString) {
		return fmt.Errorf("Only %s variable types supported", config.VariableTypeString.Printable())
	}

	def := getDefault(v)
	prompt := buildPrompt(v, def)

	rsp, err := t.cli.AskRequiredWithDefault(prompt, def)
	if err != nil {
		return err
	}

	v.Default = rsp

	delete(vs.variables, v.Name)

	return nil
}

func (vs *variables) readVars(t *module.Tree) {
	for _, v := range t.Config().Variables {
		vs.variables[fmt.Sprintf("%s.%s", t.Name(), v.Name)] = v
	}

	for _, v := range t.Children() {
		vs.readVars(v)
	}
}
