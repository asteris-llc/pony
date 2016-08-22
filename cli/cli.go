package cli

import (
	"fmt"
	"io"
	"regexp"

	"github.com/deiwin/interact"
)

type Cli struct {
	r io.Reader
	w io.Writer

	a interact.Actor
}

func New(r io.Reader, w io.Writer) *Cli {
	return &Cli{
		a: interact.NewActor(r, w),
		r: r,
		w: w,
	}
}

func (c *Cli) AskRequired(prompt string) (string, error) {
	for {
		result, err := c.a.Prompt(prompt, checkNotEmpty)
		switch {
		case isEmpty(err):
			continue
		case err != nil:
			return "", err
		}

		return result, nil
	}
}

func (c *Cli) AskRequiredWithDefault(prompt, def string) (string, error) {
	// Use AskRequired function if a default is not set
	if def == "" {
		return c.AskRequired(prompt)
	}

	result, err := c.a.PromptOptional(prompt, def)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (c *Cli) AskYesNo(prompt, def string) bool {
	result, err := c.a.PromptOptional(prompt, def)
	if err != nil {
		return false
	}

	m, err := regexp.MatchString("^[Yy][Ee]?[Ss]?$", result)
	return m
}

func (c *Cli) Printf(format string, a ...interface{}) (int, error) {
	return fmt.Fprintf(c.w, format, a...)
}

func (c *Cli) Println(a ...interface{}) (int, error) {
	return fmt.Fprintln(c.w, a...)
}
