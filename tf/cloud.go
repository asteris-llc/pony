package tf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/config/module"
)

var cloudList = map[string]string{ 
	"google": google_template,
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

func (t *Tf) SelectCloud() error {
	prompt := fmt.Sprintf("Select a cloud provider(%s)", strings.Join(cloudKeys, ","))
	for {
		rsp, err := t.cli.AskRequired(prompt)
		if err != nil {
			return err
		}

		// Trim whitespace
		rsp = strings.Trim(rsp, " 	\n")

		for c, _ := range cloudList {
			if rsp == c {
				t.cloud = c
				return nil
			}
		}
	}
}

func (t *Tf) LoadCloud() error {
	if _, ok := cloudList[t.cloud]; !ok {
		return fmt.Errorf("Cloud %s not in cloud list", t.cloud)
	}

	c, err := config.LoadJSON(bytes.NewBufferString(cloudList[t.cloud]).Bytes())
	if err != nil {
		return err
	}

	// Set a working diretory in the terraform config structure. This is required
	// if we want to on-disk modules in the future
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	c.Dir = pwd

	t.config = c
	t.m = module.NewTree("", c)
	tdir, err := tempDir()
	if err != nil {
		return err
	}

	if err := t.m.Load(&getter.FolderStorage{StorageDir: tdir}, module.GetModeUpdate); err != nil {
		return err
	}

	return nil
}

func tempDir() (string, error) {
        dir, err := ioutil.TempDir("", "tf")
        if err != nil {
		return "", err
        }
        if err := os.RemoveAll(dir); err != nil {
		return "", err
        }

        return dir, nil
}

