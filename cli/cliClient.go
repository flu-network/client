package cli

import (
	"fmt"
	"net/rpc"
	"os"
	"reflect"
	"strconv"

	"github.com/flu-network/client/common"
	transfertcp "github.com/flu-network/client/transferTCP"
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
		err = client.Call("Methods.Add", &addReq, &addResp)
		if err != nil {
			prettyPrintError(err)
		} else {
			fmt.Printf("Added numbers: %v + %v = %v.\n", addReq.A, addReq.B, addResp.Response)
		}
	case "share":
		validateArgCount("Share", ShareRequest{}, args)
		req := ShareRequest{Filepath: args[0]}
		res := ListItem{}
		err := client.Call("Methods.Share", &req, &res)
		if err != nil {
			prettyPrintError(err)
		} else {
			fmt.Print(res.Sprintf())
		}
	case "list":
		validateArgCount("List", ListRequest{}, args)
		req := ListRequest{}
		res := ListResponse{Items: []ListItem{}}
		err := client.Call("Methods.List", &req, &res)
		if err != nil {
			prettyPrintError(err)
		} else {
			for _, item := range res.Items {
				fmt.Print(item.Sprintf())
			}
		}
	// gettcp is used as a quick-and-dirty way to transfer a file over TCP to establish a benchmark.
	// It is distinct from other CLI methods in that the CLI process lives until the transfer is
	// complete.
	case "gettcp":
		validateArgCount("GetTCP", GetRequest{}, args)
		hash := common.Sha1Hash{}
		err := hash.FromStringSafe(args[0])
		validate(err)
		transfertcp.GetFile(&hash)

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

func prettyPrintError(err error) {
	if err == nil {
		panic("Cannot format a nil error")
	}
	fmt.Printf("%s\n", err)
}
