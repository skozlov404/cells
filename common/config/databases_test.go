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
			"database": "test"
		},
		"databases": {
			"test": {
				"driver": "testdriver",
				"dsn": "testdsn"
			}
		},
		"service": {
			"database": "test"
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
		driver, dsn := GetDatabase(Values("service", "database"))
		So(driver, ShouldEqual, "testdriver")
		So(dsn, ShouldEqual, "testdsn")
	})
}
