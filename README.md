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
- implement transfer! This is it!
    - Pick up by running `scripts/runDownloadTest.sh`
    - Have the receiver stream its chunks to disk
    - Have the sender maintain a set of 'unacked' messages to retransmit at the end
        - unacked messages should have a data offset and data length
    - wrap the chunk transmission up so it can run concurrently
        - Use those fancy exclusive-range things you wrote in bitset to do this
- Chunk size needs to be globally constant... 🤦‍♂️
- Use merkel trees to 'patch' the chunks if they don't match


### Local Dev notes:
- use `fn + f5` to run the client with debugging enabled in daemon mode

### Resources
- https://ops.tips/blog/udp-client-and-server-in-go/ for UDP OS-level goodies
