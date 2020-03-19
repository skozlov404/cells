package raft

import (
	"fmt"

	"github.com/micro/go-micro/registry"
	"go.etcd.io/etcd/clientv3"
)

type etcdWatcher struct {
	rch  clientv3.WatchChan
	exit chan bool
}

func (m *etcdWatcher) Next() (*registry.Result, error) {
	for wresp := range m.rch {
		for _, ev := range wresp.Events {
			fmt.Println(ev)
		}
	}

	return nil, nil
}

func (m *etcdWatcher) Stop() {
	select {
	case <-m.exit:
		return
	default:
		close(m.exit)
	}
}
