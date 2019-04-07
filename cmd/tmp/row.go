package config

func (r *RowGenerators) RunAll(ctx *Rows) {
	R := *r
	for _, x := range R {
		c := &Row{}
		x(c)
		(*ctx)[c.Name] = c
	}
}
