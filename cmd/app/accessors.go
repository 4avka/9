package app

import (
	"sort"
	"time"

	"git.parallelcoin.io/dev/9/cmd/nine"
)

// getValue returns the value contained in a Cats
func (r *Cats) getValue(cat, item string) (out *interface{}) {
	if r == nil {
		return
	} else if C, ok := (*r)[cat]; !ok {
		return
	} else if cc, ok := C[item]; !ok {
		return
	} else {
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
		return &ci
	}
}

func (r *Row) Bool() bool {
	return r.Value.Get().(bool)
}

func (r *Row) Int() int {
	return r.Value.Get().(int)
}

func (r *Row) Float() float64 {
	return r.Value.Get().(float64)
}

func (r *Row) Duration() time.Duration {
	return r.Value.Get().(time.Duration)
}

func (r *Row) Tag() string {
	return r.Value.Get().(string)
}

func (r *Row) Tags() []string {
	return r.Value.Get().([]string)
}

// GetSortedKeys returns the keys of a map in alphabetical order
func (r *CatJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *CatsJSON) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Cats) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r Cat) GetSortedKeys() (out []string) {
	for i := range r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Tokens) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}

func (r *Commands) GetSortedKeys() (out []string) {
	for i := range *r {
		out = append(out, i)
	}
	sort.Strings(out)
	return
}
