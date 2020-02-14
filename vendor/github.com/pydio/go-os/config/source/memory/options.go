package memory

import (
	"context"

	"github.com/pydio/go-os/config"
)

type changeSetKey struct{}

func withData(d []byte, f string) config.SourceOption {
	return func(o *config.SourceOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, changeSetKey{}, &config.ChangeSet{
			Data: d,
		})
	}
}

// WithChangeSet allows a changeset to be set
func WithChangeSet(cs *config.ChangeSet) config.SourceOption {
	return func(o *config.SourceOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, changeSetKey{}, cs)
	}
}

// WithJSON allows the source data to be set to json
func WithJSON(d []byte) config.SourceOption {
	return withData(d, "json")
}

// WithYAML allows the source data to be set to yaml
func WithYAML(d []byte) config.SourceOption {
	return withData(d, "yaml")
}
