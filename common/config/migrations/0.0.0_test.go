package migrations

import (
	"testing"

	"github.com/pydio/cells/common/utils/std"
	"github.com/pydio/go-os/config"
	"github.com/pydio/go-os/config/source/memory"
	. "github.com/smartystreets/goconvey/convey"
)

var (
	data = []byte(`{
		"services": {
			"pydio.api.websocket": "test"
		}
	}`)
)

func TestMigration0_0_0(t *testing.T) {
	memorySource := memory.NewSource(
		memory.WithJSON(data),
	)

	// Create new config
	c := config.NewConfig(config.WithSource(memorySource))

	var conf std.Map
	c.Get().Scan(&conf)

	Convey("Testing initial upgrade of config", t, func() {
		So(conf.Get("services/pydio.api.websocket"), ShouldNotBeNil)

		_, err := UpgradeConfigsIfRequired(conf)
		So(err, ShouldBeNil)

		PrettyPrint(conf)
		So(conf.IsEmpty(), ShouldBeFalse)

		So(conf.Get("services/pydio.api.websocket"), ShouldBeNil)
	})
}
