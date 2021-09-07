# client
Indexes, hosts and downloads files over LANs

### Run in 'daemon' mode
- `go build . && ./client -d`

### Run in 'CLI' mode
- `go build . && ./client`

### Test loopback interface
- Compile with `go build .`
- Start server `./client -d`
- Run CLI command `./client add 1 2`
