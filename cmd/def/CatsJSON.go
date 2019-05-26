package def

import "sort"

// CatsJSON is a collection of collections of lines with grouping tags
type CatsJSON map[string]CatJSON

// GetSortedKeys returns the keys of a CatsJSON in lexicographic order
func (r *CatsJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
