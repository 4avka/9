package cmd

import "testing"

func TestParseCLI(t *testing.T) {
	for i, x := range testargs {
		switch i {
		// positive tests
		case 0:
			for j, y := range x {
				if err := parseCLI(y); err != nil {
					t.Error("positive test failed", j, y, err)
				}
			}
		// negative tests
		case 1:
			for j, y := range x {
				if err := parseCLI(y); err != nil {
					t.Error("negative test failed", j, y, err)
				}
			}
		}
	}
}
