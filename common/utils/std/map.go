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
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/cast"

	"github.com/pydio/cells/common"
)

// Map structure to store configuration
type Map map[string]interface{}

// NewMap variable
func NewMap(ms ...map[string]interface{}) Map {
	if len(ms) > 0 {
		return Map(ms[0])
	}
	m := make(Map)
	return m
}

func keysToString(k ...common.Key) []string {
	var r []string

	for _, kk := range k {
		switch v := kk.(type) {
		case int:
			r = append(r, strconv.Itoa(v))
		case string:
			v = strings.Replace(v, "[", "/", -1)
			v = strings.Replace(v, "]", "/", -1)
			v = strings.Replace(v, "//", "/", -1)
			v = strings.Trim(v, "/")
			r = append(r, strings.Split(v, "/")...)
		case []string:
			for _, vv := range v {
				r = append(r, keysToString(vv)...)
			}
		}
	}

	return r
}

func (c Map) Get() interface{} {
	return c
}

func (c Map) Set(data interface{}) error {

	switch v := data.(type) {
	case []byte:
		return json.Unmarshal(v, &c)
	case map[string]interface{}:
		for k := range c {
			delete(c, k)
		}
		for k, vv := range v {
			c[k] = vv
		}
	}

	return nil
}

func (c Map) Scan(val interface{}) error {
	if c.IsEmpty() {
		return nil
	}

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

func (c Map) Values(k ...common.Key) common.ConfigValues {
	keys := keysToString(k...)

	if len(keys) == 0 {
		return c
	}

	v, ok := c[keys[0]]
	if !ok {
		return (&Value{nil, c, keys[0]}).Values(keys[1:])
	}

	if m, err := cast.ToStringMapE(v); err == nil {
		return Map(m).Values(keys[1:])
	}

	if m, ok := v.(Map); ok {
		return Map(m).Values(keys[1:])
	}

	if a, err := cast.ToSliceE(v); err == nil {
		return (&Array{a, c, keys[0]}).Values(keys[1:])
	}

	// if a, ok := v.(Array); ok {
	// 	return (&Array{a, c, keys[0]}).Values(keys[1:])
	// }

	return (&Value{v, c, keys[0]}).Values(keys[1:])
}

func (c Map) IsEmpty() bool {
	return len(c) == 0
}

// Set sets the key to value. It replaces any existing
// values.
// func (c Map) Set(key common.Key, value interface{}) error {
// 	keys := keysToString(key)
// 	pkeys := keys[0 : len(keys)-1]
// 	tkey := keys[len(keys)-1]

// 	// Retrieving existing lowest index in map
// 	var i = 0
// 	for i = 0; i < len(pkeys); i++ {
// 		if c.Get(pkeys[0:i+1]) == nil {
// 			break
// 		}
// 	}

// 	cursor := c.Get(pkeys[0:i])

// 	mcursor, ok := cursor.(Map)
// 	if !ok {
// 		return fmt.Errorf("Existing index is not a map")
// 	}

// 	// Building remaining key with an empty value
// 	var j = 0
// 	for j = i; j < len(pkeys); j++ {
// 		mcursor[pkeys[j]] = map[string]interface{}{}
// 		mcursor = mcursor[pkeys[j]].(map[string]interface{})
// 	}

// 	// Finally set the target key
// 	mcursor[tkey] = value

// 	return nil
// }

// Del deletes the values associated with key.
// func (c Map) Del(key common.Key) error {
// 	keys := keysToString(key)
// 	pkeys := keys[0 : len(keys)-1]
// 	tkey := keys[len(keys)-1]

// 	cursor := c.Get(pkeys)
// 	mcursor, ok := cursor.(map[string]interface{})
// 	if !ok {
// 		return fmt.Errorf("Existing index is not a map")
// 	}

// 	delete(mcursor, tkey)

// 	return nil
// }
