package tf

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/config/module"
)

type cloudBase struct {
	repo string
	filename string
}

var cloudList = map[string]cloudBase{ 
	"google": {
		repo: "github.com/ChrisAubuchon/pony-config/gce/base",
		filename: "gce.tf",
	},
}

var cloudKeys []string

func init() {
	cloudKeys = make([]string, len(cloudList))
	i := 0
	for k := range cloudList { 
		cloudKeys[i] = k
		i++
	}
}

func (tf *Tf) SelectCloud() error {
	prompt := fmt.Sprintf("Select a cloud provider(%s)", strings.Join(cloudKeys, ","))
	for {
		rsp, err := tf.cli.AskRequired(prompt)
		if err != nil {
			return err
		}

		for c, _ := range cloudList {
			if rsp == c {
				tf.cloud = c
				return nil
			}
		}
	}
}

func (tf *Tf) LoadCloud() error {
	if _, ok := cloudList[tf.cloud]; !ok {
		return fmt.Errorf("Cloud '%s' not in cloud list", tf.cloud)
	}

	c, err := tf.loadInternal(cloudList[tf.cloud])
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

