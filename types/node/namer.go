package node

type Namer interface {
	Name() string
}

func (n Name) Name() string    { return string(n) }
func (l Local) Name() string   { return l.VarName }
func (c Closure) Name() string { return c.VarName }
