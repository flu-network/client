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

### Test host discovery
- use scripts `runRemoteClient` and `runRemoteDaemon` in `../scripts`

### TODO:
- implement file discovery within a single host
    - i.e., "hi you, what files do you have for me?"
- implement index cleaning (i.e., remove entries for missing files)
- FTP benchmark:
    [x] setup Anna's laptop with rsync (for code) as a second client
        [ ] use `brew install inetutils`
        [ ] set up ftp servers on both machines
        [ ] this is our second benchmark 😰
    [ ] use some bash scripting to automatically run our benchmarks from personal macbook

### Local Dev notes:
- use `fn + f5` to run the client with debugging enabled in daemon mode

### Resources
- https://ops.tips/blog/udp-client-and-server-in-go/ for UDP OS-level goodies
