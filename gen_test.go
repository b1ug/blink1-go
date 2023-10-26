package blink1

import (
	"fmt"
	"sort"
	"testing"
)

func TestGetColorNames(t *testing.T) {
	var excepted = []string{
		"aqua",
		"grey",
		"fuchsia",
	}

	var names []string
	for n, c := range presetColorMap {
		ignored := false
		for _, e := range excepted {
			if n == e {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		h := convColorToHex(c)
		fmt.Printf("%q: %q,\n", h, n)
		names = append(names, n)
	}

	sort.Strings(names)
	fmt.Println()
	for _, n := range names {
		fmt.Printf("%q,", n)
	}
}
