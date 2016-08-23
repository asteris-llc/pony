package tf

import (
	"os"

	"github.com/hashicorp/terraform/config"
)

func (tf *Tf) loadInternal() (*config.Config, error) {
	c, err := config.LoadJSON(tf.cloudProvider.Root())
	if err != nil {
		return nil, err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	c.Dir = pwd

	return c, nil
}
