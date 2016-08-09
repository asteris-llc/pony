package cli

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)


func TestAskRequired_emptyLines(t *testing.T) {
	cases := []struct{
		Input string
		Expected string
	}{
		{
			"noblank\n",
			"noblank",
		},
		{
			"\nemptyline\n",
			"emptyline",
		},
		{
			"   	\nwhitespace\n",
			"whitespace",
		},
	}

	assert := assert.New(t)

	rsp := new(bytes.Buffer)
	c := New(rsp, ioutil.Discard)

	for _, cse := range cases {
		rsp.Reset()
		rsp.WriteString(cse.Input)
		result, err := c.AskRequired("Test case")
		assert.Nil(err)
		assert.Equal(cse.Expected, result)
	}
}

func Test_AskRequiredWithDefault(t *testing.T) {
	cases := []struct{
		Input string
		Default string
		Expected string
	}{
		{
			"\n",
			"test1",
			"test1",
		},
		{
			"notdefault\n",
			"default",
			"notdefault",
		},
	}

	assert := assert.New(t)

	rsp := new(bytes.Buffer)
	c := New(rsp, ioutil.Discard)

	for _, cse := range cases {
		rsp.Reset()
		rsp.WriteString(cse.Input)
		result, err := c.AskRequiredWithDefault("Test case", cse.Default)
		assert.Nil(err)
		assert.Equal(cse.Expected, result)
	}
}

func Test_AskYesNo(t *testing.T) {
	cases := []struct{
		Input string
		Expected bool
		Default string
	}{
		{ "yes\n", true, "n" },
		{ "y\n", true, "n" }, 
		{ "yE\n", true, "n" },
		{ "YES\n", true, "n" },
		{ "Yesterday\n", false, "n" },
		{ "Anything else\n", false, "n" },
		// Default value test
		{ "\n", true, "y" },
	}

	assert := assert.New(t)

	rsp := new(bytes.Buffer)
	c := New(rsp, ioutil.Discard)

	for _, cse := range cases {
		rsp.Reset()
		rsp.WriteString(cse.Input)
		result := c.AskYesNo("Test case", cse.Default)
		assert.Equal(cse.Expected, result)
	}
}
