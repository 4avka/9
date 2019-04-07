package config

type Configuration struct {
	Rows
	Name    string
	Version func() string
}

func (r *Configuration) Get(path string) interface{} {
	if _, ok := r.Rows[path]; ok {
		return r.Rows[path].Get()
	}
	return nil
}

func (r *Configuration) Put(in interface{}, path string) bool {
	if _, ok := r.Rows[path]; ok {
		return r.Rows[path].Put(in)
	}
	return false
}
