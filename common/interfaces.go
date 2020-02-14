package common

import (
	"github.com/pydio/go-os/config"
)

type Scanner interface {
	Scan(val interface{}) error
}

type Watcher interface {
	Watch() (config.Watcher, error)
}

type Key interface{}

type ConfigValues interface {
	Get() interface{}
	Set(value interface{}) error
	//Del(key Key) error

	// Bool(key Key, def ...bool) bool
	// Bytes(key Key, def ...[]byte) []byte
	// Int(key Key, def ...int) int
	// Int64(key Key, def ...int64) int64
	// Duration(key Key, def ...string) time.Duration
	// String(key Key, def ...string) string
	// StringMap(key Key) map[string]string
	// StringArray(key Key, def ...[]string) []string
	// // Map(key Key) map[string]interface{}
	// Array(key Key) Scanner
	Values(key ...Key) ConfigValues

	// Database(k Key, refs map[string]Database, def ...Database) (Database, error)

	// IsEmpty() bool

	Scanner
}

type XMLSerializableForm interface {
	Serialize(languages ...string) interface{}
}
