package config

import (
	"testing"

	"github.com/pydio/go-os/config"
	"github.com/pydio/go-os/config/source/memory"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	data = []byte(`{
		"defaults": {
			"database": {"$ref": "#/databases/test"}
		},
		"databases": {
			"test": {
				"driver": "testdriver",
				"dsn": "testdsn"
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

		db2 := Values("service1", "database").Default("#/databases/test").StringMap()
		So(db2["driver"], ShouldEqual, "testdriver")
		So(db2["dsn"], ShouldEqual, "testdsn")
	})
}
