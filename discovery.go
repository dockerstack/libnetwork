package libnetwork

import (
	"errors"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/libnetwork/config"
	"github.com/docker/swarm/discovery"
)

const DefaultHeartbeat = 10

type JoinCallback func(entries []*discovery.Entry)
type LeaveCallback func(entries []*discovery.Entry)

type HostDiscovery interface {
	StartDiscovery(*config.ClusterCfg, JoinCallback, LeaveCallback) error
	StopDiscovery() error
	Fetch() ([]*discovery.Entry, error)
}

type hostDiscovery struct {
	discovery discovery.Discovery
	nodes     []*discovery.Entry
}

func NewHostDiscovery() HostDiscovery {
	return &hostDiscovery{}
}

func (h *hostDiscovery) StartDiscovery(cfg *config.ClusterCfg, joinCallback JoinCallback, leaveCallback LeaveCallback) error {
	hb := cfg.Heartbeat
	if hb == 0 {
		hb = DefaultHeartbeat
	}
	d, err := discovery.New(cfg.Discovery, hb)
	if err != nil {
		return err
	}

	d.Watch(func(entries []*discovery.Entry) {
		processCallback(entries, joinCallback, leaveCallback)
	})

	if !checkAddrFormat(cfg.Address) {
		return errors.New("Address config should be of the form ip:port or hostname:port")
	}

	if err := d.Register(cfg.Address); err != nil {
		return err
	}

	go sustainHeartbeat(d, hb, cfg)
	return nil
}

func (h *hostDiscovery) StopDiscovery() error {
	return nil
}

func sustainHeartbeat(d discovery.Discovery, hb uint64, config *config.ClusterCfg) {
	for {
		time.Sleep(time.Duration(hb) * time.Second)
		if err := d.Register(config.Address); err != nil {
			log.Error(err)
		}
	}
}

func processCallback(entries []*discovery.Entry, joinCallback JoinCallback, leaveCallback LeaveCallback) {
}

func (h *hostDiscovery) Fetch() ([]*discovery.Entry, error) {
	if h.discovery == nil {
		return nil, errors.New("No Active Discovery")
	}
	return h.discovery.Fetch()
}

func checkAddrFormat(addr string) bool {
	m, _ := regexp.MatchString("^[0-9a-zA-Z._-]+:[0-9]{1,5}$", addr)
	return m
}
