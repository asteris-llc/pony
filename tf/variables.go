package tf

import (
	"bytes"
	"fmt"
//	"regexp"

	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/module"
	log "github.com/sirupsen/logrus"
)

const (
	_ = iota
	MetaRequired = "meta_required_variables"
	MetaIgnored = "meta_ignored_variables"

	ModuleDescription = "description"
)

var metaHandlers = []func(*Tf, *variables) error{
	metaRequiredHandler,
	metaIgnoredHandler,
}

type variables map[string]*variable

type variable struct {
	name	string
	v	*config.Variable
}

func newVariables() *variables {
	rval := new(variables)
	*rval = make(map[string]*variable)
	return rval
}

func (vs *variables) exists(name string) bool {
	_, ok := (*vs)[name]
	return ok
}

func (vs *variables) get(name string) *variable {
	if !vs.exists(name) {
		return nil
	}

	return (*vs)[name]
}

func (vs *variables) set(name string, v *variable) {
	(*vs)[name] = v
}

func (vs *variables) delete(name string) {
	delete(*vs, name)
}

func (tf *Tf) ReadVariables() error {
	vs := newVariables()
	vs.readVars(tf.tree)

	// Process the root configuration
	//
	if err := tf.processModule(tf.tree, vs, "Global Configuration"); err != nil {
		return err
	}

	// Read global variables from root. Globals will be propogated
	// through all of the sub modules
	//
	tf.globals.readVars(tf.tree)


	if err := tf.processChildren(tf.tree); err != nil {
		return err
	}

	return nil
}


func (tf *Tf) processModule(tree *module.Tree, vs *variables, header string) error {
	if header != "" {
		fmt.Printf("\n%s\n\n", header)
	}

	// Run through all of the meta variable handlers
	for _, mh := range metaHandlers {
		if err := mh(tf, vs); err != nil {
			return err
		}
	}

	// The remaining variables are optional. Run through them here
	//
	if err := optionalVariables(tf, vs); err != nil {
		return err
	}

	return nil
}

func (tf *Tf) processChildren(root *module.Tree) error {
	children := root.Children()
	for _, m := range root.Config().Modules {
		if _, ok := children[m.Name]; !ok {
			return fmt.Errorf("Module %s not found in children", m.Name)
		}

		child := children[m.Name]

		desc := ""
		if d, ok := m.RawConfig.Raw[ModuleDescription]; ok {
			switch d.(type) {
			case string:
				desc = d.(string)
			}
		}

		vs := newVariables()
		vs.readVars(child)

		for k, v := range m.RawConfig.Raw {
			switch v.(type) {
			case string:
			}

			if t := vs.get(k); t != nil {
				t.setValue(v)
					
			}
		}
		tf.processModule(child, vs, desc)


		// The RawConfig variables override the module variable's Default
		// value. We overwrite the Raw variables with whatever value the
		// the user has set. We could also delete the key from the RawConfig
		// if the two values are different.
		//
		vs = newVariables()
		vs.readVars(child)
		for k, _ := range m.RawConfig.Raw {
			if t := vs.get(k); t != nil {
				m.RawConfig.Raw[k] = t.v.Default
			}
		}
	}

	return nil
}

func optionalVariables(tf *Tf, vs *variables) error {
	if len((*vs)) == 0 {
		log.Debug("No optional variables in module")

		// No more variables left
		return nil
	}

	ask := true
	if !tf.cli.AskYesNo("Configure optional parameters? (y/N)", "n") {
		ask = false
	}

	for _, v := range *vs {
		// Set global values
		if globalValue := tf.globals.get(v.name); globalValue != nil {
			log.Debugf("Setting global value for %s", v.name)
			v.setValue(globalValue.v.Default)
		}

		if ask {
			if err := askForValue(tf, vs, v.name); err != nil {
				return err
			}
		}
	}

	return nil
}

func metaRequiredHandler(tf *Tf, vs *variables) error {
	requiredListVar := vs.get(MetaRequired)
	if requiredListVar == nil {
		// No required variables
		return nil
	}

	// Verify that the variable type is list
	if !isVariableType(requiredListVar.v, config.VariableTypeList) {
		return fmt.Errorf("Invalid type for meta_required_variables. \"%s\" != \"%s\"", 
					config.VariableTypeList, 
					requiredListVar.v.Type().Printable())
	}

	requiredList := requiredListVar.v.Default.([]interface{})
	for _, v := range requiredList {
		// variable names should only be strings
		switch v.(type) {
		case string:
		default:
			return fmt.Errorf("Invalid type for required variable name: '%T' != 'string'", v)
		}

		vname := v.(string)
		if !vs.exists(vname) {
			log.Warnf("Required variable '%s' not in module", vname)
			continue
		}

		if err := askForValue(tf, vs, vname); err != nil {
			return err
		}
	}

	// Remove meta variable from variable map
	vs.delete(MetaRequired)

	return nil
}

func metaIgnoredHandler(tf *Tf, vs *variables) error {
	ignoredListVar := vs.get(MetaIgnored)
	if ignoredListVar == nil {
		// No required variables
		return nil
	}

	// Verify that the variable type is list
	if !isVariableType(ignoredListVar.v, config.VariableTypeList) {
		return fmt.Errorf("Invalid type for meta_required_variables. \"%s\" != \"%s\"", 
					config.VariableTypeList, 
					ignoredListVar.v.Type().Printable())
	}

	ignoredList := ignoredListVar.v.Default.([]interface{})
	for _, v := range ignoredList {
		// variable names should only be strings
		switch v.(type) {
		case string:
		default:
			return fmt.Errorf("Invalid type for ignored variable name: '%T' != 'string'", v)
		}

		vname := v.(string)
		if !vs.exists(vname) {
			log.Warnf("Ignored variable '%s' not in module", vname)
			continue
		}

		vs.delete(vname)
	}

	vs.delete(MetaIgnored)

	return nil
}


func isVariableType(v *config.Variable, t config.VariableType) bool {
	return v.Type() == t
}

func (v *variable) getDefault() string {
	switch v.v.Default.(type) {
	case string:
		return v.v.Default.(string)
	}

	return ""
}

func (v *variable) buildPrompt(def string) string {
	prompt := bytes.NewBufferString("Enter a value for ")
	prompt.WriteString(v.name)

	return prompt.String()
}

func (v *variable) setValue(value interface{}) {
	v.v.Default = value
}

func askForValue(tf *Tf, vs *variables, name string) error {
	v := vs.get(name)
	if v == nil {
		return fmt.Errorf("Variable not in config: '%s'", name)
	}

	if !isVariableType(v.v, config.VariableTypeString) {
		return fmt.Errorf("Only %s variable types supported", config.VariableTypeString.Printable())
	}

	if v.v.Description != "" {
		fmt.Printf("\n%s\n", v.v.Description)
	}

	def := v.getDefault()
	prompt := v.buildPrompt(def)

	rsp, err := tf.cli.AskRequiredWithDefault(prompt, def)
	if err != nil {
		return err
	}

	v.setValue(rsp)

	vs.delete(name)

	return nil
}

func (vs *variables) readVars(t *module.Tree) {
	for _, v := range t.Config().Variables {
		if vs.exists(v.Name) {
			log.Warnf("Variable '%s' defined multiple times in module", v.Name)
		}
		vs.set(v.Name, &variable{
			name: v.Name,
			v: v,
		})
	}
}

func (tf *Tf) dumpVariables(t *module.Tree) {
	fmt.Println(t.Name())
	for _, v := range t.Config().Variables {
		fmt.Printf("variable: %s, value: %s\n", v.Name, v.Default)
	}

	for _, c := range t.Children() {
		tf.dumpVariables(c)
	}
}
