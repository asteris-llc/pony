package tf

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

func (tf *Tf) Get(dst string, u *url.URL) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	mod, err := tf.cloudProvider.GetConfig(u.Opaque)
	if err != nil {
		return err
	}

	// Hardcode the internal filename as main.tf
	//
	if err := ioutil.WriteFile(filepath.Join(dst, "main.tf"), mod, 0644); err != nil {
		return err
	}

        return nil
}

func (tf *Tf) GetFile(dst string, u *url.URL) error {
        fmt.Printf("GetFile(%s, %s)\n", dst, u.Opaque)
        return nil
}

