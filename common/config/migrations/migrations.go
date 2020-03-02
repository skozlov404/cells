package migrations

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-version"

	"github.com/pydio/cells/common"
	"github.com/pydio/cells/common/utils/migrations"
)

type migrationConfig struct {
	target *version.Version
	up     migrationConfigFunc
}

type migrationConfigFunc func(common.ConfigValues) func(context.Context) error
type migrationFunc func(common.ConfigValues) (bool, error)

var (
	configMigrations []*migrationConfig

	configKeysRenames = map[string]string{
		"services/pydio.api.websocket":            "services/" + common.SERVICE_GATEWAY_NAMESPACE_ + common.SERVICE_WEBSOCKET,
		"services/pydio.grpc.gateway.data":        "services/" + common.SERVICE_GATEWAY_DATA,
		"services/pydio.grpc.gateway.proxy":       "services/" + common.SERVICE_GATEWAY_PROXY,
		"services/pydio.rest.gateway.dav":         "services/" + common.SERVICE_GATEWAY_DAV,
		"services/pydio.rest.gateway.wopi":        "services/" + common.SERVICE_GATEWAY_WOPI,
		"ports/micro.api":                         "ports/" + common.SERVICE_MICRO_API,
		"services/micro.api":                      "services/" + common.SERVICE_MICRO_API,
		"services/pydio.api.front-plugins":        "services/" + common.SERVICE_WEB_NAMESPACE_ + common.SERVICE_FRONT_STATICS,
		"services/pydio.grpc.auth/dex/connectors": "services/" + common.SERVICE_WEB_NAMESPACE_ + common.SERVICE_OAUTH + "/connectors",
	}
	configKeysDeletes = []string{
		"services/pydio.grpc.auth/dex",
	}
)

func add(target *version.Version, m migrationConfigFunc) {
	configMigrations = append(configMigrations, &migrationConfig{target, m})
}

func getMigration(f migrationFunc) migrationConfigFunc {
	return func(c common.ConfigValues) func(context.Context) error {
		return func(context.Context) error {
			_, err := f(c)

			return err
		}
	}
}

// UpgradeConfigsIfRequired applies all registered configMigration functions
// Returns true if there was a change and save is required, error if something nasty happened
func UpgradeConfigsIfRequired(config common.ConfigValues) (bool, error) {

	v := config.Values("version")

	lastVersion, err := version.NewVersion(v.Default("0.0.0").String())
	if err != nil {
		return false, err
	}

	if !lastVersion.LessThan(common.Version()) {
		return false, nil
	}

	var mm []*migrations.Migration
	for _, m := range configMigrations {
		mm = append(mm, &migrations.Migration{
			TargetVersion: m.target,
			Up:            m.up(config),
		})
	}

	appliedVersion, err := migrations.Apply(context.Background(), lastVersion, common.Version(), mm)
	if err != nil {
		return false, err
	}

	fmt.Println("Applied version is ", appliedVersion, lastVersion)

	if !appliedVersion.GreaterThan(lastVersion) {
		return false, nil
	}

	err = v.Set(appliedVersion.String())

	return true, nil
}

// UpdateKeys replace a key with a new one
func UpdateKeys(config common.ConfigValues, m map[string]string) (bool, error) {
	var save bool
	for oldPath, newPath := range m {
		val := config.Values(oldPath)
		if val != nil {
			fmt.Printf("[Configs] Upgrading: renaming key %s to %s\n", oldPath, newPath)
			config.Values(newPath).Set(val.Get())
			val.Del()
			save = true
		}
	}
	return save, nil
}

// UpdateVals replace a val with a new one
func UpdateVals(config common.ConfigValues, m map[string]string) (bool, error) {

	var all interface{}
	err := config.Scan(&all)
	if err != nil {
		return false, err
	}

	var save bool
	all = parseAndReplace(all, func(a map[string]interface{}) map[string]interface{} {
		for oldV, newV := range m {
			for k, v := range a {
				if vv, ok := v.(string); ok && vv == oldV {
					fmt.Printf("[Configs] Upgrading: renaming val %s to %s\n", oldV, newV)
					a[k] = newV
					save = true
				}
			}
		}

		return a
	})

	if !save {
		return save, nil
	}

	config.Set(all)

	return true, nil
}

func deleteConfigKeys(config common.ConfigValues) (bool, error) {
	var save bool
	for _, oldPath := range configKeysDeletes {
		val := config.Values(oldPath)
		var data interface{}
		if e := val.Scan(&data); e == nil && data != nil {
			fmt.Printf("[Configs] Upgrading: deleting key %s\n", oldPath)
			val.Del()
			save = true
		}
	}
	return save, nil
}

// dsnRemoveAllowNativePassword removes this part from default DSN
// func dsnRemoveAllowNativePassword(config *Config) (bool, error) {
// 	testFile := filepath.Join(ApplicationWorkingDir(ApplicationDirServices), common.SERVICE_GRPC_NAMESPACE_+common.SERVICE_CONFIG, "version")
// 	if data, e := ioutil.ReadFile(testFile); e == nil && len(data) > 0 {
// 		ref, _ := version.NewVersion("1.4.1")
// 		if v, e2 := version.NewVersion(strings.TrimSpace(string(data))); e2 == nil && v.LessThan(ref) {
// 			dbId := config.Get("defaults", "database").String("")
// 			if dbId != "" {
// 				if dsn := config.Get("databases", dbId, "dsn").String(""); dsn != "" && strings.Contains(dsn, "allowNativePasswords=false\u0026") {
// 					dsn = strings.Replace(dsn, "allowNativePasswords=false\u0026", "", 1)
// 					fmt.Println("[Configs] Upgrading DSN to support new MySQL authentication plugin")
// 					config.Set(dsn, "databases", dbId, "dsn")
// 					return true, nil
// 				}
// 			}
// 		}
// 	}
// 	return false, nil
// }

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

type ReplacerFunc func(map[string]interface{}) map[string]interface{}

func parseAndReplace(i interface{}, replacer ReplacerFunc) interface{} {

	switch m := i.(type) {
	case []map[string]interface{}:
		var new []map[string]interface{}
		for _, mm := range m {
			new = append(new, parseAndReplace(mm, replacer).(map[string]interface{}))
		}

		return new
	case map[string]interface{}:
		new := replacer(m)
		for k, v := range new {
			new[k] = parseAndReplace(v, replacer)
		}
		return new
	case []interface{}:
		var new []interface{}
		for _, mm := range m {
			new = append(new, parseAndReplace(mm, replacer))
		}

		return new

	}

	return i
}
