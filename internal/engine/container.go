package engine

type Container struct {
	services map[string]any
}

func NewContainer() *Container {
	return &Container{
		services: make(map[string]any),
	}
}

func (c *Container) Register(name string, svc any) {
	c.services[name] = svc
}

func (c *Container) Get(name string) any {
	return c.services[name]
}
