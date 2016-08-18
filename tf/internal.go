package tf

import (
	"path/filepath"
	"os"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/terraform/config"
)

func (tf *Tf) loadInternal() (*config.Config, error) {
	if err := getter.Get(tf.tempDir, tf.cloudProvider.Repo()); err != nil {
		return nil, err
	}

	path := filepath.Join(tf.tempDir, tf.cloudProvider.Base())

	c, err := config.LoadFile(path)
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
