package main

import (
	"flag"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/cli"
)

// Defaults. Can be overridden by user
var catalogueDir = "/usr/local/var/flu-network/catalogue" // TODO: make cross-platform
var sockaddr = "/tmp/flu-network.sock"

func main() {
	daemonMode := flag.Bool("d", false, "-d")
	flag.Parse()

	if *daemonMode {
		startDaemon()
	} else {
		argsWithoutProg := os.Args[1:]
		enterCli(argsWithoutProg)
	}
}

func enterCli(args []string) {
	cliClient := cli.NewClient(sockaddr)
	cliClient.Run(args)
}

func startDaemon() {
	cat := &catalogue.Cat{DataDir: catalogueDir}
	verify(cat.Init())
	/*
		TODO: set up harnessing: e.g.,
			- handle OS signals properly
			- set up channels for control signals between goroutines
	*/

	// Expose CLI interface
	go func() {
		err := os.RemoveAll(sockaddr)
		verify(err)

		addr, err := net.ResolveUnixAddr("unixgram", sockaddr)
		verify(err)

		rpcServer := rpc.NewServer()
		cliMethods := cli.NewMethods(cat)
		rpcServer.Register(cliMethods)
		listener, e := net.ListenUnix("unix", addr)
		verify(e)

		rpcServer.Accept(listener)
		listener.Close()
	}()

	go func() {
		// TODO: Expose p2p interface
	}()

	for {
		// TODO: Download files that need downloading
		time.Sleep(time.Millisecond * 1000)
	}
}

func verify(err error) {
	if err != nil {
		panic(err)
	}
}
