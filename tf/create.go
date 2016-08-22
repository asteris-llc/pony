package tf

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
)

var create_metaHandlers = []metaHandler{
	create_metaProviderHandler,
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

func create_metaProviderHandler(tf *Tf, vs *variables) error {
	providerList, err := vs.getStringList(MetaProvider)
	if err != nil {
		return err
	}

	if providerList == nil {
		return nil
	}

	providerVars, err := tf.cloudProvider.GetProviderVars()
	if err != nil {
		return err
	}

	for _, vname := range providerList {
		pv := vs.get(vname)
		if pv == nil {
			return fmt.Errorf("Provider variable '%s' not defined in module", vname)
			continue
		}
		if providerVar, ok := providerVars[vname]; ok {
			pv.setValue(providerVar)
		} else {
			if err := askForValue(tf, vs, vname); err != nil {
				return err
			}
		}
	}

	return nil
}

func create_metaRequiredHandler(tf *Tf, vs *variables) error {
	requiredList, err := vs.getStringList(MetaRequired)
	if err != nil {
		return err
	}

	if requiredList == nil {
		return nil
	}

	for _, vname := range requiredList {
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
	ignoredList, err := vs.getStringList(MetaIgnored)
	if err != nil {
		return err
	}

	if ignoredList == nil {
		return nil
	}

	for _, vname := range ignoredList {
		if !vs.exists(vname) {
			log.Warnf("Ignored variable '%s' not in module", vname)
			continue
		}

		vs.delete(vname)
	}

	vs.delete(MetaIgnored)

	return nil
}
