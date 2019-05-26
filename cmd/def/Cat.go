package def

import "sort"

// Cat is a collection of Rows with tag labels
type Cat map[string]*Row

// CatGenerator is a function that configures a Cat
type CatGenerator func(ctx *Cat)

// CatGenerators is a collection of Cat's
type CatGenerators []CatGenerator


func (r Cat) GetSortedKeys() (out []string) {
	for i := range r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
