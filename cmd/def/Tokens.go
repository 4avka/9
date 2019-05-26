package def

import "sort"

// Token is a struct that ties together CLI invocation to the Command it
// relates to
type Token struct {
	Value string
	Cmd   Command
}

// Tokens is a collection of Tokens
type Tokens map[string]Token

// Optional is a set of possible valid items accompanying a Token
type Optional []string

// Precedent is a set of possible valid items that match preferentially
// to the item in a Command
type Precedent []string

// GetSortedKeys returns a slice of Tokens keys in lexicographic order
func (r *Tokens) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
