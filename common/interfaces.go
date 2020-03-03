package common

import (
	"time"

	"github.com/pydio/go-os/config"
)

type Scanner interface {
	Scan(val interface{}) error
}

type Watcher interface {
	Watch() (config.Watcher, error)
}

type Key interface{}

type ConfigValue interface {
	Default(interface{}) ConfigValue

	Bool() bool
	// Bytes() []byte
	Int() int
	Int64() int64
	Duration() time.Duration
	String() string
	StringMap() map[string]string
	StringArray() []string
	Slice() []interface{}
	Map() map[string]interface{}
	// Database() Database

	// Scanner
}

type ConfigValues interface {
	Get() ConfigValue
	Set(value interface{}) error
	Del() error
	Values(key ...Key) ConfigValues

	ConfigValue

	Scanner
}

type XMLSerializableForm interface {
	Serialize(languages ...string) interface{}
}
