package main

import (
	"flag"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"time"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/cli"
	transfertcp "github.com/flu-network/client/transferTCP"
)

// Defaults. Can be overridden by user
var catalogueDir = "/usr/local/var/flu-network/catalogue" // TODO: make cross-platform
var sockaddr = "/tmp/flu-network.sock"                    // for cli communication
var tcpPort = 17969                                       // port 'f1' in hex

func main() {
	daemonMode := flag.Bool("d", false, "-d")
	flag.Parse()

	if *daemonMode {
		startDaemon()
	} else {
		args := os.Args[1:] // first arg is pathToBinary. Should be ignored in a CLI.
		// cliClient is designed to be a short-lived process that executes a single CLI command,
		// waits for the result, prints it and then exits.
		cli.NewClient(sockaddr).Run(args)
	}
}

func startDaemon() {
	cat, err := catalogue.NewCat(catalogueDir)
	failHard(err)
	failHard(cat.Init())
	/*
		TODO: set up harnessing: e.g.,
			- handle OS signals properly
			- set up channels for control signals between goroutines
	*/

	// Expose CLI interface
	go func() {
		err := os.RemoveAll(sockaddr)
		failHard(err)

		addr, err := net.ResolveUnixAddr("unixgram", sockaddr)
		failHard(err)

		rpcServer := rpc.NewServer()
		cliMethods := cli.NewMethods(cat)
		rpcServer.Register(cliMethods)
		listener, e := net.ListenUnix("unix", addr)
		failHard(e)

		rpcServer.Accept(listener)
		listener.Close()
	}()

	go func() {
		// TODO: Expose p2p interface
	}()

	// expose TCP interface. Only for benchmarking.
	go func() {
		server, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", tcpPort))
		failHard(err)
		defer server.Close()
		fmt.Println("TCP Server started. Waiting for connections...")
		for {
			connection, err := server.Accept()
			failHard(err)
			go transfertcp.SendFile(connection, cat)
		}
	}()

	for {
		// TODO: Download files that need downloading
		time.Sleep(time.Millisecond * 1000)
	}
}

func failHard(err error) {
	if err != nil {
		panic(err)
	}
}
