package cli

import (
	"fmt"
	"strconv"
)

func (c *Cli) Select(varName string, list []string) (string, error) {
	count := len(list)

	if count <= 0 {
		return "", fmt.Errorf("Selecting from an empty list")
	}

	prompt := fmt.Sprintf("Enter value for %s (1-%d)", varName, count)

	for {
		c.Println()
		for i, item := range list {
			c.Printf("%2d %s\n", i+1, item)
		}

		result, err := c.a.Prompt(prompt, checkIsNumeric, checkNotEmpty)
		switch {
		case isEmpty(err):
			fallthrough
		case isNotNumeric(err):
			continue
		case err != nil:
			return "", err
		}

		// String conversion errors are caught by the checkIsNumeric check
		rval, _ := strconv.Atoi(result)

		if (rval < 1) || (rval > count) {
			continue
		}

		return list[rval-1], nil
	}
}

func (c *Cli) SelectMany(varName string, list []string) ([]string, error) {
	count := len(list)
	if count <= 0 {
		return nil, fmt.Errorf("Selecting from an empty list")
	}

	selected := make([]bool, count)

	prompt := fmt.Sprintf("Enter value for %s (1-%d)", varName, count+1)

	for {
		c.Println()
		for i, item := range list {
			marked := " "
			if selected[i] {
				marked = "*"
			}
			c.Printf("%s %2d %s\n", marked, i+1, item)
		}
		c.Printf("  %2d Done\n", count+1)

		result, err := c.a.Prompt(prompt, checkIsNumeric, checkNotEmpty)
		switch {
		case isEmpty(err):
			fallthrough
		case isNotNumeric(err):
			continue
		case err != nil:
			return nil, err
		}

		// String conversion errors are caught by the checkIsNumeric check
		rval, _ := strconv.Atoi(result)

		if rval == (count + 1) {
			rlist := []string{}
			for i, isSelected := range selected {
				if isSelected {
					rlist = append(rlist, list[i])
				}
			}
			if len(rlist) <= 0 {
				continue
			}

			return rlist, nil
		}

		if (rval < 1) || (rval > count) {
			continue
		}

		selected[rval-1] = !selected[rval-1]
	}
}
