package common

import (
	"log"
	"log/slog"
	"net"
)

// DnsDiscovery returns all the peers behin d an fqdn name.
// The current implementation works only in a  Kubernetes cluster, and the onloy supported fqdn is a Kubernetes Headless Service
// An headless service fqdn is identified as <service-name>.<namespace>.svc
// @see https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
type DnsDiscovery struct {
	Fqdn string
}

func (d *DnsDiscovery) DiscoverPeers() ([]string, error) {
	if !IsKube() {
		slog.Info("distributed-map: not running in a Kube cluster, returning no peers")
		return []string{}, nil
	}

	ips, err := net.LookupIP(d.Fqdn)
	if err != nil {
		slog.Error("distributed-map: failed to lookup peers for fqdn, returning no peers", "error", err, "fqdn", d.Fqdn)
		return []string{}, nil
	}

	peers := make([]string, len(ips))
	for i, ip := range ips {
		peers[i] = ip.String()
		slog.Info("distributed-map: discovered peer", "ip", ip)
	}

	return peers, nil
}

func (d *DnsDiscovery) Initialize() error {
	// println("==== DnsDiscovery: initialize ====")
	return nil
}

func (d *DnsDiscovery) SetLogger(_ *log.Logger) {
	// println("==== DnsDiscovery: SetLogger ====")
}

func (d *DnsDiscovery) SetConfig(_ map[string]interface{}) error {
	// println("==== DnsDiscovery: SetConfig ====")
	// for k, v := range cfg {
	// 	println(fmt.Sprintf("    -> %v : %v", k, v))
	// }
	return nil
}

func (d *DnsDiscovery) Register() error {
	// println("==== DnsDiscovery: Register ====")
	return nil
}

func (d *DnsDiscovery) Deregister() error {
	// println("==== DnsDiscovery: Deregister ====")
	return nil
}

func (d *DnsDiscovery) Close() error {
	// println("==== DnsDiscovery: Close ====")
	return nil
}
