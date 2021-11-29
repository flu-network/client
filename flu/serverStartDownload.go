package flu

import (
	"crypto/sha1"
	"fmt"

	"github.com/flu-network/client/common"
)

// StartDownload creates a progressfile for the specified file, adds it to the catalogue, and
// begins the download. A name for the file is chosen arbitrarily from one of the hosts who have
// that file.
func (s *Server) StartDownload(hash *common.Sha1Hash) error {
	extantRecord, _ := s.cat.Contains(hash)

	if extantRecord == nil {
		fmt.Println("Download not registered. TODO: register download later.")
	}

	// TOOD: MAKE CONCURRENT BY RUNNING IN GOROUTINES
	// list hosts that have info on this file
	hostResponses := s.DiscoverHosts(hash, make([]uint16, 0))

	for _, hostResponse := range hostResponses {
		// skip hosts that know nothing about the file we want
		if len(hostResponse.Chunks) == 0 {
			continue
		}

		fmt.Println(hostResponse)

		ip, port := hostResponse.Address, hostResponse.Port
		key := downloadKey{
			hash:       *hash,
			remoteHost: hostResponse.Address,
		}

		if _, ok := s.downloads[key]; !ok {
			conn, err := DialPeer(ip, port, hash, 0)
			if err != nil {
				panic(err) // TODO: log somewhere and move on...
			}
			s.downloads[key] = conn

			for { // returns false when done or errored
				packet, ok := conn.Read() // blocks execution
				if !ok {
					break
				}

				if packet.Offset == 0 {
					// By convention the 0-offset packet contains the hash and chunk size
					hash, size, data := packet.Split()
					conn.hash = hash
					conn.buffer = make([]byte, size)
					copy(conn.buffer, data) // start at zero offset
					conn.bytesReceived += len(data)
				} else {
					conn.bytesReceived += len(packet.Data)
					if len(packet.Data) > 0 {
						copy(conn.buffer[packet.Offset:], packet.Data)
					} else {
						hash := sha1.New()
						hash.Write(conn.buffer)
						finalHash := (&common.Sha1Hash{}).FromSlice(hash.Sum(nil))
						if finalHash.Data != conn.hash.Data {
							// TODO: Handle this a little more gracefully...
							panic("Sha1 hashes did not match. Download was corrupted")
						}
						break
					}
				}
				fmt.Printf("Recieved %d/%d\n", conn.bytesReceived, len(conn.buffer))
				conn.Ack(packet.Offset)
			}
		}
	}

	// // register the download in the catalogue, so its persisted
	// err := s.cat.RegisterDownload(
	// 	entry.SizeInBytes,
	// 	entry.ChunkCount,
	// 	entry.ChunkSizeInBytes,
	// 	entry.Sha1Hash,
	// 	entry.FileName,
	// )
	// if err != nil {
	// 	return err
	// }
	// }
	return nil
}
