package client

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"time"

	log "github.com/Sirupsen/logrus"

	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/swarm/discovery"
	"github.com/docker/swarm/discovery/token"
)

// CmdNetworkCreate handles Network Create UI
func (cli *NetworkCli) CmdNetworkCreate(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "create", "NETWORK-NAME", chain+" create", false)
	flDriver := cmd.String([]string{"d", "-driver"}, "", "Driver to manage the Network")
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	if *flDriver == "" {
		*flDriver = "bridge"
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("POST", "/networks/"+args[0], nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkRm handles Network Delete UI
func (cli *NetworkCli) CmdNetworkRm(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "rm", "NETWORK-NAME", chain+" rm", false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("DELETE", "/networks/"+args[0], nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkLs handles Network List UI
func (cli *NetworkCli) CmdNetworkLs(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "ls", "", chain+" ls", false)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("GET", "/networks", nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkInfo handles Network Info UI
func (cli *NetworkCli) CmdNetworkInfo(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "info", "NETWORK-NAME", chain+" info", false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("GET", "/networks/"+args[0], nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkJoin handles the UI to let a Container join a Network via an endpoint
// Sample UI : <chain> network join <container-name/id> <network-name/id> [<endpoint-name>]
func (cli *NetworkCli) CmdNetworkJoin(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "join", "CONTAINER-NAME/ID NETWORK-NAME/ID [ENDPOINT-NAME]",
		chain+" join", false)
	cmd.Require(flag.Min, 2)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("POST", "/endpoints/", nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkLeave handles the UI to let a Container disconnect from a Network
// Sample UI : <chain> network leave <container-name/id> <network-name/id>
func (cli *NetworkCli) CmdNetworkLeave(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "leave", "CONTAINER-NAME/ID NETWORK-NAME/ID",
		chain+" leave", false)
	cmd.Require(flag.Min, 2)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO : Proper Backend handling
	obj, _, err := readBody(cli.call("PUT", "/endpoints/", nil, nil))
	if err != nil {
		fmt.Fprintf(cli.err, "%s", err.Error())
		return err
	}
	if _, err := io.Copy(cli.out, bytes.NewReader(obj)); err != nil {
		return err
	}
	return nil
}

// CmdNetworkClusterCreate Create new Cluster
func (cli *NetworkCli) CmdNetworkClusterCreate(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "cluster create", "", chain+" cluster create", false)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO: Add proper backend handling.
	// Temporarily integrating Discovery code directly in the UI path for testing purposes
	discovery := &token.Discovery{}
	discovery.Initialize("", 0)
	token, err := discovery.CreateCluster()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)
	return nil
}

// CmdNetworkClusterLs Lists cluster nodes
func (cli *NetworkCli) CmdNetworkClusterLs(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "cluster ls", "<discovery>", chain+" cluster ls", false)
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO: Add proper backend handling.
	// Temporarily integrating Discovery code directly in the UI path for testing purposes
	d, err := discovery.New(args[0], 0)
	if err != nil {
		log.Fatal(err)
	}

	nodes, err := d.Fetch()
	if err != nil {
		log.Fatal(err)
	}
	for _, node := range nodes {
		fmt.Println(node)
	}
	return nil
}

// CmdNetworkClusterJoin Join an existing Cluster
func (cli *NetworkCli) CmdNetworkClusterJoin(chain string, args ...string) error {
	cmd := cli.Subcmd(chain, "cluster join", "<discovery>", chain+" cluster join", false)
	flAddr := cmd.String([]string{"-a", "-addr"}, "", "ip to advertise")
	cmd.Require(flag.Min, 1)
	err := cmd.ParseFlags(args, true)
	if err != nil {
		return err
	}
	// TODO: Add proper backend handling.
	// Temporarily integrating Discovery code directly in the UI path for testing purposes
	d, err := discovery.New(args[1], 0)
	if err != nil {
		log.Fatal(err)
	}

	if !checkAddrFormat(*flAddr) {
		log.Fatal("--addr should be of the form ip:port or hostname:port")
	}

	if err := d.Register(*flAddr); err != nil {
		log.Fatal(err)
	}

	for {
		log.Infof("Registering on the discovery service every 3 seconds... %s %s",
			args[0], *flAddr)
		time.Sleep(3 * time.Second)
		if err := d.Register(*flAddr); err != nil {
			log.Error(err)
		}
	}
	//return nil
}

func checkAddrFormat(addr string) bool {
	m, _ := regexp.MatchString("^[0-9a-zA-Z._-]+:[0-9]{1,5}$", addr)
	return m
}
