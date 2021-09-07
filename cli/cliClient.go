package cli

import (
	"fmt"
	"net/rpc"
	"os"
	"reflect"
	"strconv"
)

// Client is the entrypoint for code that runs on the CLI process. The user types commands
// in here and they get executed by the CLI methods on the daemon process that hosts the CLI
// methods
type Client struct {
	sockaddr string
}

// NewClient returns a new client instance
func NewClient(addr string) *Client {
	return &Client{
		sockaddr: addr,
	}
}

// Run is the entry point for the CLI
func (c *Client) Run(cmdArgs []string) {
	if len(cmdArgs) == 0 {
		fmt.Println("No command supplied") // TODO: print something useful
		return
	}

	client, err := rpc.Dial("unix", c.sockaddr)
	verify(err)

	cmd := cmdArgs[0]
	args := cmdArgs[1:]

	switch cmd {
	case "add":
		validateArgCount("Add", AddRequest{}, args)
		a, err := strconv.Atoi(args[0])
		validate(err)
		b, err := strconv.Atoi(args[1])
		validate(err)
		addReq := AddRequest{A: a, B: b}
		addResp := AddResponse{}
		client.Call("Methods.Add", &addReq, &addResp)
		fmt.Printf("Added numbers: %v + %v = %v.\n", addReq.A, addReq.B, addResp.Response)
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func validateArgCount(method string, request interface{}, args []string) {
	required := reflect.TypeOf(request).NumField()
	if required != len(args) {
		fmt.Printf("%s Expects %d arguments for but got %d\n", method, required, len(args))
		os.Exit(2)
	}
}

func validate(err error) {
	if err != nil {
		// handle error
		fmt.Println(err)
		os.Exit(2)
	}
}

func verify(err error) {
	if err != nil {
		panic(err)
	}
}
