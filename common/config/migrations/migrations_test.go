package migrations

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pydio/cells/common/utils/std"
	"github.com/pydio/go-os/config"
	"github.com/pydio/go-os/config/source/memory"
	. "github.com/smartystreets/goconvey/convey"
)

func TestMigrations(t *testing.T) {
	data := []byte(`{}`)

	memorySource := memory.NewSource(
		memory.WithJSON(data),
	)

	// Create new config
	c := config.NewConfig(config.WithSource(memorySource))

	var conf std.Map
	c.Get().Scan(&conf)

	Convey("Testing initial upgrade of config", t, func() {
		_, err := UpgradeConfigsIfRequired(conf)
		So(err, ShouldBeNil)

		So(conf.IsEmpty(), ShouldBeFalse)
	})
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}
