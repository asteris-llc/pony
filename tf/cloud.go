package tf

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/config/module"
)

func (tf *Tf) SelectCloud() error {
	prompt := fmt.Sprintf("Select a cloud provider(%s)", strings.Join(tf.cloudList.Keys(), ","))
	for {
		rsp, err := tf.cli.AskRequired(prompt)
		if err != nil {
			return err
		}

		if found := tf.cloudList.GetProvider(rsp); found != nil {
			tf.cloudProvider = found
			return nil
		}
	}
}

func (tf *Tf) LoadCloud() error {
	c, err := tf.loadInternal()
	if err != nil {
		return err
	}

	tf.tree = module.NewTree("", c)
	if err != nil {
		return err
	}

	if err := tf.tree.Load(&getter.FolderStorage{StorageDir: tf.tempDir}, module.GetModeUpdate); err != nil {
		return err
	}

	return nil
}

