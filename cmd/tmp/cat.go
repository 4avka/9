package config

type Cat map[string]Row

type CatGenerator func(ctx *Cat)

type CatGenerators []CatGenerator

func (r *CatGenerators) RunAll(cat Cat) {
	for _, x := range *r {
		x(&cat)
	}
	return
}
