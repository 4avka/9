package def

import (
	"sort"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
)

// Cats are a collection of Rows with a string tag
type Cats map[string]Cat

// GetSortedKeys returns the keys in a Cats map in lexicographic order
func (r *Cats) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

// getValue returns the value contained in a Cats
func (r *Cats) getValue(cat, item string) (out *interface{}) {
	if r == nil {
		return
	} else if C, ok := (*r)[cat]; !ok {
		return
	} else if cc, ok := C[item]; !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		o := cc.Value.Get()
		return &o
	}
}

// Str returns the pointer to a value in the category map
func (r *Cats) Str(cat, item string) (out *string) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(string); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Tags returns the pointer to a value in the category map
func (r *Cats) Tags(cat, item string) (out *[]string) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.([]string); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Map returns the pointer to a value in the category map
func (r *Cats) Map(cat, item string) (out *nine.Mapstringstring) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(nine.Mapstringstring); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Int returns the pointer to a value in the category map
func (r *Cats) Int(cat, item string) (out *int) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(int); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Bool returns the pointer to a value in the category map
func (r *Cats) Bool(cat, item string) (out *bool) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(bool); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Float returns the pointer to a value in the category map
func (r *Cats) Float(cat, item string) (out *float64) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(float64); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// Duration returns the pointer to a value in the category map
func (r *Cats) Duration(cat, item string) (out *time.Duration) {
	cv := r.getValue(cat, item)
	if cv == nil {
		return
	}
	CC := *cv
	if ci, ok := CC.(time.Duration); !ok {
		return
	} else {
		// Ignore linter, this return value is in if/else block scope
		return &ci
	}
}

// RunAll runs a collection of CatGenerators on a Cat
func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}
