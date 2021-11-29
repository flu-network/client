package cli

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	"reflect"

	"github.com/flu-network/client/common"
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
		fmt.Println("No command supplied")
		return
	}

	client, err := rpc.Dial("unix", c.sockaddr)
	verify(err)

	cmd := cmdArgs[0]
	args := cmdArgs[1:]

	switch cmd {
	// Share makes a file available on the flu network. Share assumes that the file you give it has
	// already been downloaded in its entirety and will never change. If the content of the file
	// changes later, flu will raise an error when an integrity check is next performed.
	// Usage:
	// 	- flu list ~/Desktop/path-to-file.mkv
	case "share":
		validateArgCount("Share", ShareRequest{}, args)
		req := ShareRequest{Filepath: args[0]}
		res := ListItem{}
		callClientMethodAndPrintResponse(client, "Methods.Share", &req, &res)

	// Clean checks the integrity of the local flu index. Specifically it:
	// - Removes missing files from the index
	// - Removes files that have sha1 hashes that do not match the indexed sha1 hash
	// Usage:
	// 	- flu clean
	case "clean":
		validateArgCount("Clean", CleanRequest{}, args)
		req := CleanRequest{}
		res := CleanResponse{}
		callClientMethodAndPrintResponse(client, "Methods.Clean", &req, &res)

	// List lists the files availble for download. If an IP address is supplied, it lists the files
	// available on the node at that IP address. If not, it lists the files available on the local
	// daemon. Right now, it only supports IPV4 addresses.
	// Usage:
	// 	- flu list 					# list files indexed on local daemon
	// 	- flu list 192.168.86.39 	# list files on 192.168.86,39
	case "list":
		req := ListRequest{
			IP:       []byte{},
			Sha1Hash: (&common.Sha1Hash{}).Blank(),
		}
		res := ListResponse{Items: []ListItem{}}
		if len(args) > 0 {
			addr := net.ParseIP(args[0])
			if addr == nil {
				prettyPrintError(fmt.Errorf("invalid IP Address: %s", args[0]))
				return
			}
			req.IP = addr
		}
		if len(args) > 1 {
			err := req.Sha1Hash.FromStringSafe(args[1])
			validate(err)
		}
		callClientMethodAndPrintResponse(client, "Methods.List", &req, &res)

	// Get starts downloading the specified file from all available hosts. If a transfer has already
	// been started Get does not affect it. Get runs in the background so will run whenever the flu
	// daemon is running until the file is downloaded. Get implicitly also shares the file that is
	// being downloaded.
	// Usage:
	//   - flu get A0F1490A20D0211C997B44BC357E1972DEAB8AE3 # get file with this sha1 hash
	//   - flu get A0F1490A20D0211C997B44BC357E1972DEAB8AE3 --sercet # get this file and don't share
	case "get":
		req := GetRequest{
			Sha1Hash: &common.Sha1Hash{},
		}
		res := GetResponse{}
		err := req.Sha1Hash.FromStringSafe(args[0])
		validate(err)
		validateArgCount("Clean", GetRequest{}, args)
		callClientMethodAndPrintResponse(client, "Methods.Get", &req, &res)

	// Chims lists available hosts on the LAN, including the local daemon. If gives hosts a few
	// seconds to responds and then prints the response from all hosts that replied.
	// Usage:
	// 	- flu chims # list available hosts
	case "chims":
		req := ChimRequest{}
		res := ChimResponseList{}
		hash := common.Sha1Hash{}
		if len(args) > 0 {
			err := hash.FromStringSafe(args[0])
			validate(err)
			req.Sha1Hash = &hash
		} else {
			req.Sha1Hash = hash.Blank()
		}
		callClientMethodAndPrintResponse(client, "Methods.Chims", &req, &res)

	default:
		fmt.Printf("Unknown command: %s\n", cmd)
	}
}

func callClientMethodAndPrintResponse(c *rpc.Client, m string, req interface{}, res Printable) {
	err := c.Call(m, req, res)
	if err != nil {
		prettyPrintError(err)
	} else {
		fmt.Print(res.Sprintf())
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
