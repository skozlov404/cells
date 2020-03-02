package std

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pydio/cells/common"
)

// Value is standard
type Value struct {
	v interface{}
	p interface{} // Reference to parent for assignment
	k interface{} // Reference to key for re-assignment
}

// Get retrieve interface
func (v *Value) Get() common.ConfigValue {
	if v == nil || v.v == nil {
		return nil
	}
	return &def{v.v}
}

// Default value set
func (v *Value) Default(i interface{}) common.ConfigValue {
	vv := v.Get()
	if vv == nil {
		return &def{nil}
	}
	return vv.Default(i)
}

// Set data in interface
func (v *Value) Set(data interface{}) error {
	if v == nil {
		return fmt.Errorf("Value doesn't exist")
	}
	if m, ok := v.p.(*Map); ok {
		m.v[v.k.(string)] = data
	}
	if a, ok := v.p.(*Array); ok {
		old := a.Get().Slice()
		old[v.k.(int)] = data
		a.Set(old)
	}

	v.v = data
	return nil
}

func (v *Value) Del() error {
	if v == nil {
		return fmt.Errorf("Value doesn't exist")
	}
	if m, ok := v.p.(*Map); ok {
		delete(m.v, v.k.(string))
	}
	if a, ok := v.p.(*Array); ok {
		old := a.Get().Slice()
		idx := v.k.(int)
		old = append(old[:idx], old[idx+1:]...)
		a.Set(old)
	}

	v.v = nil
	return nil
}

// Values cannot retrieve lower values as it is final
func (v *Value) Values(k ...common.Key) common.ConfigValues {
	keys := keysToString(k...)

	// A value arriving there with another key below if of the wrong type
	if len(keys) > 0 {
		return (*Value)(nil)
	}

	return v
}

// Scan to interface
func (v *Value) Scan(val interface{}) error {
	jsonStr, err := json.Marshal(v.v)
	if err != nil {
		return err
	}

	switch v := val.(type) {
	case proto.Message:
		err = (&jsonpb.Unmarshaler{AllowUnknownFields: true}).Unmarshal(bytes.NewReader(jsonStr), v)
	default:
		err = json.Unmarshal(jsonStr, v)
	}

	return err
}

func (c *Value) Bool() bool {
	return c.Default(false).Bool()
}
func (c *Value) Int() int {
	return c.Default(0).Int()
}
func (c *Value) Int64() int64 {
	return c.Default(0).Int64()
}
func (c *Value) Duration() time.Duration {
	return c.Default(0 * time.Second).Duration()
}
func (c *Value) String() string {
	return c.Default("").String()
}
func (c *Value) StringMap() map[string]string {
	return c.Default(map[string]string{}).StringMap()
}
func (c *Value) StringArray() []string {
	return c.Default([]string{}).StringArray()
}
func (c *Value) Slice() []interface{} {
	return c.Default([]interface{}{}).Slice()
}
func (c *Value) Map() map[string]interface{} {
	return c.Default(map[string]interface{}{}).Map()
}
