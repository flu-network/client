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
	"github.com/flu-network/client/flu"
)

// Defaults. Can be overridden by user
const catalogueDir = "/usr/local/var/flu-network/catalogue" // TODO: make cross-platform
const sockaddr = "/tmp/flu-network.sock"                    // for cli communication
const tcpPort = 61697                                       // port 'f101' in hex
const udpPort = 61696                                       // port "f100" in hex

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
	fluServer := flu.NewServer(udpPort, cat)
	/*
		TODO: set up harnessing: e.g.,
			- handle OS signals properly
			- set up channels for control signals between goroutines (e.g., to stop a transfer)
	*/

	// Expose CLI interface (RPC over unix domain sockets)
	go func() {
		err := os.RemoveAll(sockaddr)
		failHard(err)

		addr, err := net.ResolveUnixAddr("unixgram", sockaddr)
		failHard(err)

		rpcServer := rpc.NewServer()
		cliMethods := cli.NewMethods(cat, fluServer)
		rpcServer.Register(cliMethods)
		listener, e := net.ListenUnix("unix", addr)
		failHard(e)
		fmt.Printf("UNIX Interface available at: %s\n", sockaddr)

		rpcServer.Accept(listener)
		listener.Close()
	}()

	// expose p2p interface (UDP)
	go func() {
		addr := net.UDPAddr{IP: nil, Port: udpPort, Zone: ""}
		c1, err := net.ListenUDP("udp", &addr)
		failHard(err)
		defer c1.Close()
		fmt.Printf("UDP Interface available at: %s:%d\n", fluServer.LocalIP().String(), udpPort)

		for {
			buffer := make([]byte, 1024)
			_, returnAddress, err := c1.ReadFromUDP(buffer)
			failHard(err)
			go func() {
				err := fluServer.HandleMessage(buffer, c1, returnAddress)
				if err != nil {
					fmt.Println(err)
				}
			}()
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
