package config

type Cats map[string]Cat

type CatsGenerator func(ctx *Cats)

type CatsGenerators []CatsGenerator

func (r *CatsGenerators) RunAll(cats Cats) {
	for _, x := range *r {
		x(&cats)
	}
	return
}
