package common

import (
	"context"
	"github.com/buraksezer/olric"
	"github.com/buraksezer/olric/config"
	"sync"
	"time"
)

// NewDistributedMap return an olric.DMap, which is an map-like struct replicated on all nodes of a cluster.
// clusterFqdn is an fqdn address that resolve the ip addresses of the cluster node.
// All the keys whithin the same namespace are managed by the same dmap: 2 dmap with the same namespace
// have the same values.
// The size of the dmap is bounded to 500MB.
// @see https://github.com/buraksezer/olric
func NewDistributedMap(_ context.Context, clusterFqdn string, namespace string) (olric.DMap, error) {

	cfg := config.New("lan")
	cfg.ReplicationMode = config.AsyncReplicationMode
	cfg.ReplicaCount = 2 // copy everything to 2 nodes
	cfg.ReadRepair = true
	cfg.ServiceDiscovery = map[string]interface{}{
		"plugin": &DnsDiscovery{
			Fqdn: clusterFqdn,
		},
	}
	cfg.DMaps.EvictionPolicy = config.LRUEviction
	cfg.DMaps.MaxInuse = 500_000_000 // 500 MB
	cfg.BootstrapTimeout = 2 * time.Minute
	cfg.MemberCountQuorum = 1

	// keep logging level high while we
	// test and use the dmap for a while
	cfg.LogLevel = "DEBUG"
	cfg.LogVerbosity = 6

	// this wait group is used to block the main goroutine until the embedded client is ready
	wg := sync.WaitGroup{}
	wg.Add(1)
	cfg.Started = func() {
		wg.Done()
	}

	cache, err := olric.New(cfg)
	if err != nil {
		return nil, err
	}

	// start the instance, which triggers the k8s service discovery and forms the cluster
	go func() {
		err = cache.Start()
		if err != nil {
			panic(err)
		}
	}()

	// wait for the cluster to be ready before continuing
	wg.Wait()

	embClient := cache.NewEmbeddedClient()
	dmap, err := embClient.NewDMap(namespace)
	return dmap, err
}
