package main

import (
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/libnetwork/client"
)

func callback(method, path string, data interface{}, headers map[string][]string) (io.ReadCloser, int, error) {
	return nil, 0, nil
}

func main() {
	if len(os.Args) < 3 {
		printUsage()
		os.Exit(1)
	}
	_, stdout, stderr := term.StdStreams()
	cli := client.NewNetworkCli(stdout, stderr, callback)
	if err := cli.Cmd("dnet", os.Args[1:]...); err != nil {
		logrus.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("Usage: dnet network <subcommand> <OPTIONS>")
}
