package def

import "sort"

// Line is the JSON formatted version of a Cat
type Line struct {
	Value   interface{} `json:"value"`
	Default interface{} `json:"default,omitempty"`
	Min     int         `json:"min,omitempty"`
	Max     int         `json:"max,omitempty"`
	Usage   string      `json:"usage"`
}

// CatJSON is a collection of lines with their tag
type CatJSON map[string]Line

// GetSortedKeys returns the keys of a map in alphabetical order
func (r *CatJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
