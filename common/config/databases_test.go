package config

import (
	"testing"

	"github.com/pydio/cells/common/utils/std"
	"github.com/pydio/go-os/config"
	"github.com/pydio/go-os/config/source/memory"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	data = []byte(`{
		"defaults": {
			"database": {"$ref": "#/databases/test3"}
		},
		"databases": {
			"test": {
				"driver": "testdriver",
				"dsn": "testdsn"
			},
			"test2": {
				"driver": "testdriver2",
				"dsn": "testdsn2"
			},
			"test3": {
				"driver": "testdriver3",
				"dsn": "testdsn3"
			}
		},
		"service1": {
			"database": {"$ref": "#/databases/test"}
		},
		"service2": {
		}
	}`)
)

func TestDatabase(t *testing.T) {
	memorySource := memory.NewSource(
		memory.WithJSON(data),
	)

	// Create new config
	c := config.NewConfig(config.WithSource(memorySource))

	defaultConfig = &Config{c}
	once.Do(func() {})

	Convey("Testing initial upgrade of config", t, func() {
		db := Values("service1", "database").StringMap()
		So(db["driver"], ShouldEqual, "testdriver")
		So(db["dsn"], ShouldEqual, "testdsn")

		db2 := Values("service2", "database").Default(std.Reference("#/databases/test2")).StringMap()
		So(db2["driver"], ShouldEqual, "testdriver2")
		So(db2["dsn"], ShouldEqual, "testdsn2")

		db3 := Values("service2", "database").Default(std.Reference("#/defaults/database")).StringMap()
		So(db3["driver"], ShouldEqual, "testdriver3")
		So(db3["dsn"], ShouldEqual, "testdsn3")
	})
}
