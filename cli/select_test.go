package cli

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {

	cases := []struct {
		Options  []string
		Input    string
		Expected string
	}{
		{
			[]string{"Chris", "Jenny", "Ethan", "Emily"},
			"1\n",
			"Chris",
		},
		{
			[]string{"Chris", "Jenny", "Ethan", "Emily"},
			"44\n2\n",
			"Jenny",
		},
		{
			[]string{"Chris", "Jenny", "Ethan", "Emily"},
			"\n4\n",
			"Emily",
		},
		{
			[]string{"Chris", "Jenny", "Ethan", "Emily"},
			"-4\n3\n",
			"Ethan",
		},
		{
			[]string{"Chris", "Jenny", "Ethan", "Emily"},
			"Ethan\n3\n",
			"Ethan",
		},
	}

	assert := assert.New(t)

	rsp := new(bytes.Buffer)
	c := New(rsp, ioutil.Discard)

	for _, cse := range cases {
		rsp.Reset()
		rsp.WriteString(cse.Input)
		result, err := c.Select("test", cse.Options)
		assert.Nil(err)
		assert.Equal(cse.Expected, result)
	}
}
