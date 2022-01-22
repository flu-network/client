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
-- IMMEDIATE CONCERNS --
- Make downloads happen in parallel. Fix the massive hack in serverStartDownload.go
- Make downloads start automatically when the daemon restarts
- Make downloads cancellable (using 'context'?)

-- General ugliness -- 
- downloading a file overwrites (without first deleting) the extant file at that path in ~/Downloads
  . This leads to some hella confusing behavior. Delete the file first!
- Universally replace []uint16 with []range wherever possible
- Have the sender maintain a set of 'unacked' messages to retransmit at the end
- Chunk size needs to be globally constant... 🤦‍♂️
- Use merkel trees to 'patch' the chunks if they don't match
- have the receiver close the connection instead of *relying* on a sender-side timeout


### Local Dev notes:
- use `fn + f5` to run the client with debugging enabled in daemon mode

### Resources
- https://ops.tips/blog/udp-client-and-server-in-go/ for UDP OS-level goodies
