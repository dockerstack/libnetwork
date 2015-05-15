package hostdiscovery

import (
	"errors"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"

	mapset "github.com/deckarep/golang-set"
	"github.com/docker/libnetwork/config"
	"github.com/docker/swarm/discovery"
	// Annonymous import will be removed after we upgrade to latest swarm
	_ "github.com/docker/swarm/discovery/file"
	// Annonymous import will be removed after we upgrade to latest swarm
	_ "github.com/docker/swarm/discovery/kv"
	// Annonymous import will be removed after we upgrade to latest swarm
	_ "github.com/docker/swarm/discovery/nodes"
	// Annonymous import will be removed after we upgrade to latest swarm
	_ "github.com/docker/swarm/discovery/token"
)

const defaultHeartbeat = 10

// JoinCallback provides a callback event for new node joining the cluster
type JoinCallback func(entries []net.IP)

// LeaveCallback provides a callback event for node leaving the cluster
type LeaveCallback func(entries []net.IP)

// HostDiscovery primary interface
type HostDiscovery interface {
	StartDiscovery(*config.ClusterCfg, JoinCallback, LeaveCallback) error
	StopDiscovery() error
	Fetch() ([]net.IP, error)
}

type hostDiscovery struct {
	discovery discovery.Discovery
	nodes     []net.IP
	stopChan  chan struct{}
	sync.Mutex
}

// NewHostDiscovery function creates a host discovery object
func NewHostDiscovery() HostDiscovery {
	return &hostDiscovery{nodes: []net.IP{}, stopChan: make(chan struct{})}
}

func (h *hostDiscovery) StartDiscovery(cfg *config.ClusterCfg, joinCallback JoinCallback, leaveCallback LeaveCallback) error {
	hb := cfg.Heartbeat
	if hb == 0 {
		hb = defaultHeartbeat
	}
	d, err := discovery.New(cfg.Discovery, hb)
	if err != nil {
		return err
	}

	go d.Watch(func(entries []*discovery.Entry) {
		h.processCallback(entries, joinCallback, leaveCallback)
	})

	if ip := net.ParseIP(cfg.Address); ip == nil {
		return errors.New("Address config should be either ipv4 or ipv6 address")
	}

	if err := d.Register(cfg.Address + ":0"); err != nil {
		return err
	}

	h.Lock()
	h.discovery = d
	h.Unlock()

	go sustainHeartbeat(d, hb, cfg, h.stopChan)
	return nil
}

func (h *hostDiscovery) StopDiscovery() error {
	h.Lock()
	stopChan := h.stopChan
	h.Unlock()

	close(stopChan)
	return nil
}

func sustainHeartbeat(d discovery.Discovery, hb uint64, config *config.ClusterCfg, stopChan chan struct{}) {
	for {
		time.Sleep(time.Duration(hb) * time.Second)

		select {
		case <-stopChan:
			return
		default:
		}

		if err := d.Register(config.Address + ":0"); err != nil {
			log.Error(err)
		}
	}
}

func (h *hostDiscovery) processCallback(entries []*discovery.Entry, joinCallback JoinCallback, leaveCallback LeaveCallback) {
	updated := hosts(entries)
	h.Lock()
	existing := h.nodes
	h.Unlock()
	added, removed := diff(existing, updated)
	if len(added) > 0 {
		joinCallback(added)
	}
	if len(removed) > 0 {
		leaveCallback(removed)
	}
}

func diff(existing []net.IP, updated []net.IP) (added []net.IP, removed []net.IP) {
	updateSet := mapset.NewSet()
	for _, ip := range updated {
		updateSet.Add(ip.String())
	}
	existingSet := mapset.NewSet()
	for _, ip := range existing {
		existingSet.Add(ip.String())
	}
	addSlice := updateSet.Difference(existingSet).ToSlice()
	removeSlice := existingSet.Difference(updateSet).ToSlice()
	for _, ip := range addSlice {
		added = append(added, net.ParseIP(ip.(string)))
	}
	for _, ip := range removeSlice {
		removed = append(removed, net.ParseIP(ip.(string)))
	}
	return
}

func (h *hostDiscovery) Fetch() ([]net.IP, error) {
	h.Lock()
	hd := h.discovery
	h.Unlock()
	if hd == nil {
		return nil, errors.New("No Active Discovery")
	}
	entries, err := hd.Fetch()
	if err != nil {
		return nil, err
	}
	return hosts(entries), nil
}

func hosts(entries []*discovery.Entry) []net.IP {
	hosts := []net.IP{}
	for _, entry := range entries {
		hosts = append(hosts, net.ParseIP(entry.Host))
	}
	return hosts
}
