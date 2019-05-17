package ifc

// Iface is an abstract container that simplifies handling interface
// variables
type Iface struct {
	Data *interface{}
}

// NewIface returns a new Iface to the caller loaded with an empty
// interface variable
func NewIface() *Iface {
	return &Iface{Data: new(interface{})}
}

// Get returns the dereferenced iface container
func (i *Iface) Get() interface{} {
	if i == nil {
		return nil
	}
	if i.Data == nil {
		return nil
	}
	return *i.Data
}

// Put loads an Iface with the given variable
func (i *Iface) Put(in interface{}) *Iface {
	if i == nil {
		i = NewIface()
	}
	if i.Data == nil {
		i.Data = new(interface{})
	}
	*i.Data = in
	return i
}
