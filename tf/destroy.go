package tf

import (
	"bytes"
	"fmt"
	"io/ioutil"

	"github.com/hashicorp/terraform/terraform"

	log "github.com/sirupsen/logrus"
)

var destroy_metaHandlers = []metaHandler{
	destroy_metaDestroyHandler,
}

func (tf *Tf) Destroy(statePath string) error {
	data, err := ioutil.ReadFile(statePath)
	if err != nil {
		return err
	}

	state, err := terraform.ReadState(bytes.NewReader(data))
	if err != nil {
		return err
	}
	tf.state = state

	// Get the outputs of the root module
	root := tf.state.RootModule()

	// Check for a "cloud" output telling us what configuration
	// to use
	if cloudVar, ok := root.Outputs["cloud"]; !ok {
		if err := tf.SelectCloud(); err != nil {
			return err
		}
	} else {
		// Verify that the cloud output is a string
		switch cloudVar.Value.(type) {
		case string:
			cloudName := root.Outputs["cloud"].Value.(string)
			if cp := tf.cloudList.GetProvider(cloudName); cp != nil {
				tf.cloudProvider = cp
				break
			}

			// Unknown cloud. Ask
			if err := tf.SelectCloud(); err != nil {
				return err
			}
		default:
			if err := tf.SelectCloud(); err != nil {
				return err
			}
		}
	}

	// Load the cloud configuration
	if err := tf.LoadCloud(); err != nil {
		return err
	}

	// Read the configuration variables
	if err := tf.ReadVariables(destroy_metaHandlers); err != nil {
		return err
	}

	if err := tf.Context(true); err != nil {
		return err
	}

	if err := tf.Plan(); err != nil {
		return err
	}

	if err := tf.Apply(); err != nil {
		return err
	}

	return nil
}

func destroy_metaDestroyHandler(tf *Tf, vs *variables) error {
	destroyList, err := vs.getStringList(MetaDestroy)
	if err != nil {
		return err
	}

	if destroyList == nil {
		fmt.Println("Nothing to do")
		return nil
	}

	root := tf.state.RootModule()

	for _, vname := range destroyList {
		if !vs.exists(vname) {
			log.Warnf("Required variable '%s' not in module", vname)
			continue
		}

		if outputVar, ok := root.Outputs[vname]; ok {
			// Only strings allowed
			if outputVar.Type != "string" {
				continue
			}

			dvar := vs.get(vname)
			dvar.setValue(outputVar.Value.(string))
		}

		if err := askForValue(tf, vs, vname); err != nil {
			return err
		}
	}

	vs.delete(MetaDestroy)

	return nil
}
