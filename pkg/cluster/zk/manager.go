package zk

func (c *controller) RegisterResource(input string, resource string) error {
	return c.zc.CreatePersistent(c.kb.resource(resource), []byte(input))
}

func (c *controller) RegisteredResources() (map[string][]string, error) {
	resources, inputs, err := c.zc.ChildrenValues(c.kb.resources())
	if err != nil {
		return nil, err
	}

	r := make(map[string][]string)
	for i, encodedResource := range resources {
		res, _ := c.kb.decodeResource(encodedResource)
		input := string(inputs[i])
		if _, present := r[input]; !present {
			r[input] = []string{res}
		} else {
			r[input] = append(r[input], res)
		}
	}

	return r, nil
}

func (c *controller) Open() error {
	return c.connectToZookeeper()
}

func (c *controller) Close() {
	c.zc.Disconnect()
}
