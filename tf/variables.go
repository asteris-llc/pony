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
	MetaDestroy = "meta_destroy_variables"

	ModuleDescription = "description"
)

type metaHandler func(*Tf, *variables) error

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

func (tf *Tf) ReadVariables(mh []metaHandler) error {
	vs := newVariables()
	vs.readVars(tf.tree)

	// Process the root configuration
	//
	if err := tf.processModule(tf.tree, vs, mh, "Global Configuration"); err != nil {
		return err
	}

	// Read global variables from root. Globals will be propogated
	// through all of the sub modules
	//
	tf.globals.readVars(tf.tree)


	if err := tf.processChildren(tf.tree, mh); err != nil {
		return err
	}

	return nil
}


func (tf *Tf) processModule(tree *module.Tree, vs *variables, mh []metaHandler, header string) error {
	if header != "" {
		fmt.Printf("\n%s\n\n", header)
	}

	// Run through all of the meta variable handlers
	for _, m := range mh {
		if err := m(tf, vs); err != nil {
			return err
		}
	}

	return nil
}

func (tf *Tf) processChildren(root *module.Tree, mh []metaHandler) error {
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
		tf.processModule(child, vs, mh, desc)


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
