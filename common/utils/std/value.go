package std

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pydio/cells/common"
)

type Value struct {
	v interface{}
	p interface{} // Reference to parent for assignment
	k interface{} // Reference to key for re-assignment
}

// Get retrieve interface
func (v *Value) Get() interface{} {
	if v == nil {
		return nil
	}
	if m, ok := v.p.(Map); ok {
		return m[v.k.(string)]
	}
	if a, ok := v.p.(Array); ok {
		return a[v.k.(int)]
	}
	return v.v
}

func (v *Value) Set(data interface{}) error {
	if v == nil {
		return fmt.Errorf("Value doesn't exist")
	}
	if m, ok := v.p.(Map); ok {
		m[v.k.(string)] = data
	}
	fmt.Println("v.p ", reflect.TypeOf(v.p), v.k, data)
	if a, ok := v.p.(Array); ok {
		a[v.k.(int)] = data
		fmt.Println(a)
	}

	fmt.Println("HERE ", v)
	v.v = data
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

// func (v *Value) String() string {

// 	return ""
// }
