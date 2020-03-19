// Package raft provides a raft registry
package raft

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/cmd"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/server"
	"github.com/ory/viper"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/embed"
	"go.etcd.io/etcd/raft"

	defaults "github.com/pydio/cells/common/micro"
)

func init() {
	// STARTING SERVERS
	go func() {
		lpurl, _ := url.Parse(viper.GetString("registry_address"))
		lcurl, _ := url.Parse(viper.GetString("registry_cluster_address"))

		cfg := embed.NewConfig()
		cfg.Dir = "default.etcd"
		cfg.LPUrls[0] = 
		e, err := embed.StartEtcd(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer e.Close()
		select {
		case <-e.Server.ReadyNotify():
			log.Printf("Server is ready!")
		case <-time.After(60 * time.Second):
			e.Server.Stop() // trigger a shutdown
			log.Printf("Server took too long to start!")
		}
		log.Fatal(<-e.Err())
	}()
}

type raftRegistry struct {
	c *clientv3.Client
}

var (
	storage = raft.NewMemoryStorage()
)

func init() {
	cmd.DefaultRegistries["raft"] = NewRegistry
}

func (m *raftRegistry) watch(r *registry.Result) {
}

func (m *raftRegistry) Options() registry.Options {
	return registry.Options{}
}

func (m *raftRegistry) GetService(service string) ([]*registry.Service, error) {
	fmt.Println("Getting service ", service)
	return nil, nil
}

func (m *raftRegistry) ListServices() ([]*registry.Service, error) {
	resp, err := m.c.Get(context.Background(), "services_", clientv3.WithPrefix())
	fmt.Println(resp, err)

	return []*registry.Service{}, nil
}

func (m *raftRegistry) Register(s *registry.Service, opts ...registry.RegisterOption) error {

	for _, node := range s.Nodes {
		b, err := json.Marshal(node)
		if err != nil {
			return err
		}

		m.c.Put(context.Background(), "services_"+node.Id, string(b))
	}

	return nil
}

func (m *raftRegistry) Deregister(s *registry.Service) error {
	return nil
}

func (m *raftRegistry) Watch(opts ...registry.WatchOption) (registry.Watcher, error) {
	rch := m.c.Watch(context.Background(), "services_", clientv3.WithPrefix())

	return &etcdWatcher{
		rch:  rch,
		exit: make(chan bool),
	}, nil
}

func (m *raftRegistry) String() string {
	return "raft"
}

func NewRegistry(opts ...registry.Option) registry.Registry {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
	}
	// defer cli.Close()

	return &raftRegistry{c: cli}
}

func Enable() {
	addr := viper.GetString("registry_address")
	r := NewRegistry(
		registry.Addrs(addr),
		registry.Timeout(1*time.Second),
	)

	defaults.InitServer(func() server.Option {
		return server.Registry(r)
	})

	defaults.InitClient(func() client.Option {
		return client.Registry(r)
	})

	registry.DefaultRegistry = r

}
