// Package memory is a memory source
package memory

import (
	"crypto/md5"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pydio/go-os/config"
)

type memory struct {
	sync.RWMutex
	ChangeSet *config.ChangeSet
	Watchers  map[string]*watcher
}

func (s *memory) Read() (*config.ChangeSet, error) {
	s.RLock()
	cs := &config.ChangeSet{
		Timestamp: s.ChangeSet.Timestamp,
		Data:      s.ChangeSet.Data,
		Checksum:  s.ChangeSet.Checksum,
		Source:    s.ChangeSet.Source,
	}
	s.RUnlock()
	return cs, nil
}

func (s *memory) Watch() (config.SourceWatcher, error) {
	w := &watcher{
		Id:      uuid.New().String(),
		Updates: make(chan *config.ChangeSet, 100),
		Source:  s,
	}

	s.Lock()
	s.Watchers[w.Id] = w
	s.Unlock()
	return w, nil
}

func (m *memory) Write(cs *config.ChangeSet) error {
	return nil
}

// Update allows manual updates of the config data.
func (s *memory) Update(c *config.ChangeSet) {
	// don't process nil
	if c == nil {
		return
	}

	// hash the file
	s.Lock()
	// update changeset
	s.ChangeSet = &config.ChangeSet{
		Data:      c.Data,
		Source:    "memory",
		Timestamp: time.Now(),
	}
	// hash the file
	h := md5.New()
	h.Write(c.Data)
	checksum := fmt.Sprintf("%x", h.Sum(nil))

	s.ChangeSet.Checksum = checksum

	// update watchers
	for _, w := range s.Watchers {
		select {
		case w.Updates <- s.ChangeSet:
		default:
		}
	}
	s.Unlock()
}

func (s *memory) String() string {
	return "memory"
}

func NewSource(opts ...config.SourceOption) config.Source {
	var options config.SourceOptions
	for _, o := range opts {
		o(&options)
	}

	s := &memory{
		Watchers: make(map[string]*watcher),
	}

	if options.Context != nil {
		c, ok := options.Context.Value(changeSetKey{}).(*config.ChangeSet)
		if ok {
			s.Update(c)
		}
	}

	return s
}
