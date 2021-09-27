# client
Indexes, hosts and downloads files over LANs

### Run in 'daemon' mode
- `go build . && ./client -d`

### Run in 'CLI' mode
- `go build . && ./client`

### Run all tests
- `go test ./...`

### Test loopback interface
- Compile with `go build .`
- Start server `./client -d`
- Run CLI command `./client add 1 2`

### Test Indexing a file
- `go build . && ./client -d`
- `cat /usr/share/dict/words > /tmp/testFile.txt`
- `go build . && ./client share /tmp/testFile.txt`

### Test Listing files
- `go build . && ./client -d`
- `go build . && ./client list`

### TODO:
- implement `flu tcpget host sha1hash`
    [x] use port binary.BigEndian.Uint16([]byte{"F", "0"}) as UDP port
    [x] use port binary.BigEndian.Uint16([]byte{"F", "1"}) as TCP port
    [x] add a TCP listener inside main(), waiting for msg containing sha1 hash
        [x] on receipt just send the whole file
        [x] use reference: https://mrwaggel.be/post/golang-transfer-a-file-over-a-tcp-socket/
        - this is our first benchmark 😨

- Get test script working (~/Documents/code/bradfield/csi/flu/client/scripts/hostDiscovery.sh)
    [ ] compile
    [ ] copy binary
    [ ] clear remote index
    [ ] rebuild remote index
    [ ] start remote daemon
    [ ] start local daemon
    [ ] run `flu chims` successfully (see below first)


- implement host discovery so we can do tcpget over LAN instead of just localhost
    [ ] complete CLI method chims
        - CLI -> daemon (via RPC) -> remote hosts (via UDP)
        - reference https://ops.tips/blog/udp-client-and-server-in-go/ for UDP OS-level goodies

- Second FTP benchmark:
    [x] setup Anna's laptop with rsync (for code) as a second client
        [ ] use `brew install inetutils`
        [ ] set up ftp servers on both machines
        [ ] this is our second benchmark 😰
    [ ] use some bash scripting to automatically run our benchmarks from personal macbook

### Local Dev notes:
- use `fn + f5` to run the client with debugging enabled in daemon mode
