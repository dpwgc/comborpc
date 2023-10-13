package comborpc

import (
	"encoding/json"
	"encoding/xml"
	"gopkg.in/yaml.v3"
)

// Next
// go to the next processing method
func (c *Context) Next() {
	c.index++
	for c.index < len(c.methods) {
		c.methods[c.index](c)
		c.index++
	}
}

// Abort
// stop continuing down execution
func (c *Context) Abort() {
	c.index = len(c.methods) + 1
}

func (c *Context) ReadString() string {
	return c.input
}

func (c *Context) ReadJson(v any) error {
	return json.Unmarshal([]byte(c.input), v)
}

func (c *Context) ReadYaml(v any) error {
	return yaml.Unmarshal([]byte(c.input), v)
}

func (c *Context) ReadXml(v any) error {
	return xml.Unmarshal([]byte(c.input), v)
}

func (c *Context) WriteString(data string) {
	c.output = data
}

func (c *Context) WriteJson(v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}

func (c *Context) WriteYaml(v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}

func (c *Context) WriteXml(v any) error {
	data, err := xml.Marshal(v)
	if err != nil {
		return err
	}
	c.output = string(data)
	return nil
}
