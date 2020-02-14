package std

import (
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	data = []byte(`{
		"service": {
			"val": "test",
			"map": {
				"val": "test"
			},
			"array": [1,2,3,4],
			"arrayMap": [{
				"val": "test",
				"map": {
					"val": "test"
				}
			}]
		}
	}`)
)

func TestStd(t *testing.T) {
	Convey("Testing map get", t, func() {
		var m Map
		err := json.Unmarshal(data, &m)
		So(err, ShouldBeNil)

		So(m.Values("service").Get(), ShouldNotBeNil)
		So(m.Values("fakeservice").Get(), ShouldBeNil)

		So(m.Values("service/val").Get(), ShouldEqual, "test")
		So(m.Values("service", "val").Get(), ShouldEqual, "test")
		So(m.Values("service", "fakeval").Get(), ShouldBeNil)

		So(m.Values("service", "array"), ShouldNotBeNil)
		So(m.Values("service", "array", "1").Get(), ShouldEqual, 2)
		So(m.Values("service", "array", 1).Get(), ShouldEqual, 2)
		So(m.Values("service", "array", 1, 2).Get(), ShouldBeNil)

		So(m.Values("service/array[1]").Get(), ShouldEqual, 2)
		So(m.Values("service/array[1][2]").Get(), ShouldBeNil)
		So(m.Values("service/array[1][2]").Get(), ShouldBeNil)

		So(m.Values("service/arrayMap[0]/val").Get(), ShouldEqual, "test")
		So(m.Values("service/arrayMap[0]/fakeval").Get(), ShouldBeNil)
		So(m.Values("service/arrayMap[1]/val").Get(), ShouldBeNil)
		So(m.Values("service/arrayMap[0]/map/val").Get(), ShouldEqual, "test")
		So(m.Values("service/arrayMap[0]/map[val]").Get(), ShouldEqual, "test")
	})

	Convey("Testing map full set", t, func() {
		m := NewMap()

		err := m.Set(data)
		So(err, ShouldBeNil)

		So(m.Values("service").Get(), ShouldNotBeNil)
		So(m.Values("fakeservice").Get(), ShouldBeNil)

		So(m.Values("service/val").Get(), ShouldEqual, "test")
		So(m.Values("service", "val").Get(), ShouldEqual, "test")
		So(m.Values("service", "fakeval").Get(), ShouldBeNil)

		So(m.Values("service", "array"), ShouldNotBeNil)
		So(m.Values("service", "array", "1").Get(), ShouldEqual, 2)
		So(m.Values("service", "array", 1).Get(), ShouldEqual, 2)
		So(m.Values("service", "array", 1, 2).Get(), ShouldBeNil)

		So(m.Values("service/array[1]").Get(), ShouldEqual, 2)
		So(m.Values("service/array[1][2]").Get(), ShouldBeNil)
		So(m.Values("service/array[1][2]").Get(), ShouldBeNil)

		So(m.Values("service/arrayMap[0]/val").Get(), ShouldEqual, "test")
		So(m.Values("service/arrayMap[0]/fakeval").Get(), ShouldBeNil)
		So(m.Values("service/arrayMap[1]/val").Get(), ShouldBeNil)
		So(m.Values("service/arrayMap[0]/map/val").Get(), ShouldEqual, "test")
		So(m.Values("service/arrayMap[0]/map[val]").Get(), ShouldEqual, "test")
	})

	Convey("Testing replacing a string value", t, func() {
		m := NewMap()
		m.Set(data)

		// Replacing a value
		err := m.Values("service", "map").Set(map[string]interface{}{
			"val2": "test2",
		})
		So(err, ShouldBeNil)

		So(m.Values("service", "fakemap", "val").Set("test"), ShouldNotBeNil) // Should throw an error
		So(m.Values("service", "map", "val2").Set("test3"), ShouldBeNil)      // Should not throw an error
		So(m.Values("service", "map", "val2").Get(), ShouldEqual, "test3")

		So(m.Values("service", "map2").Set(NewMap()), ShouldBeNil)
		So(m.Values("service", "map2", "val").Set("test"), ShouldBeNil)
		So(m.Values("service", "map2", "val").Get(), ShouldEqual, "test")
		So(m.Values("service", "array2").Set(make([]interface{}, 2)), ShouldBeNil)
		So(m.Values("service", "array2", "val").Set("test"), ShouldNotBeNil) // Array should have int index
		So(m.Values("service", "array2", "0").Set("test"), ShouldBeNil)      // Array should have int index
		fmt.Println("Array ", m.Values("service", "array2"))
		So(m.Values("service", "array2", "0").Get(), ShouldEqual, "test")
	})
}
