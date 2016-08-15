package tf

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/config"

	log "github.com/sirupsen/logrus"
)

var create_metaHandlers = []metaHandler{
	create_metaRequiredHandler,
	create_metaIgnoredHandler,
	create_metaOptionalHandler,
}

func (tf *Tf) Create() error {
	if err := tf.LoadCloud(); err != nil {
		return err
	}

	if err := tf.ReadVariables(create_metaHandlers); err != nil {
		return err
	}

	if err := tf.Context(false); err != nil {
		return err
	}

//	tf.dumpVariables(tf.tree)

	if err := tf.Plan(); err != nil {
		return err
	}

	if err := tf.Apply(); err != nil {
		return err
	}

	return nil
}

func create_metaOptionalHandler(tf *Tf, vs *variables) error {
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
		// Skip unhandled meta variables
		if strings.HasPrefix(v.name, "meta_") {
			continue
		}

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

func create_metaRequiredHandler(tf *Tf, vs *variables) error {
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

func create_metaIgnoredHandler(tf *Tf, vs *variables) error {
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

