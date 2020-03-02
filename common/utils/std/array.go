/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package std

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pydio/cells/common"
	"github.com/spf13/cast"
)

type Array struct {
	v []interface{}
	p interface{} // Reference to parent for assignment
	k interface{} // Reference to key for re-assignment
}

func (c *Array) Get() common.ConfigValue {
	return &def{c.v}
}

func (c *Array) Default(i interface{}) common.ConfigValue {
	return c.Get().Default(i)
}

func (c *Array) Set(v interface{}) error {
	if c == nil {
		return fmt.Errorf("Value doesn't exist")
	}
	if m, ok := c.p.(*Map); ok {
		m.v[c.k.(string)] = v
	}

	c.v = v.([]interface{})
	return nil
}

func (c *Array) Del() error {
	if c == nil {
		return fmt.Errorf("Value doesn't exist")
	}
	if m, ok := c.p.(*Map); ok {
		delete(m.v, c.k.(string))
	}

	c.v = nil
	return nil
}

func (c *Array) Values(k ...common.Key) common.ConfigValues {
	keys := keysToString(k...)

	if len(keys) == 0 {
		return c
	}

	// Specific handling for pointers
	if keys[0] == "#" {
		if c.p != nil {
			if v, ok := c.p.(common.ConfigValues); ok {
				return v.Values(keys)
			}
		}
		return c.Values(keys[1:])
	}

	idx, err := cast.ToIntE(keys[0])
	if err != nil {
		return (*Value)(nil)
	}

	if len(c.v) <= idx {
		return (*Value)(nil)
	}

	v := c.v[idx]

	keys = keys[1:]

	if m, err := cast.ToStringMapE(v); err == nil {
		return (&Map{m, c, idx}).Values(keys)
	}

	return (&Value{v, c, idx}).Values(keys)
}

func (c *Array) Scan(val interface{}) error {
	jsonStr, err := json.Marshal(c)
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

func (c *Array) Bool() bool {
	return c.Default(false).Bool()
}
func (c *Array) Int() int {
	return c.Default(0).Int()
}
func (c *Array) Int64() int64 {
	return c.Default(0).Int64()
}
func (c *Array) Duration() time.Duration {
	return c.Default(0 * time.Second).Duration()
}
func (c *Array) String() string {
	return c.Default("").String()
}
func (c *Array) StringMap() map[string]string {
	return c.Default(map[string]string{}).StringMap()
}
func (c *Array) StringArray() []string {
	return c.Default([]string{}).StringArray()
}
func (c *Array) Slice() []interface{} {
	return c.Default([]interface{}{}).Slice()
}
func (c *Array) Map() map[string]interface{} {
	return c.Default(map[string]interface{}{}).Map()
}
