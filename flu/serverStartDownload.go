package flu

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

// StartDownload creates a progressfile for the specified file, adds it to the catalogue, and
// begins the download. A name for the file is chosen arbitrarily from one of the hosts who have
// that file.
func (s *Server) StartDownload(hash *common.Sha1Hash) error {
	extantRecord, _ := s.cat.Contains(hash)

	ownIP := s.LocalIP()
	ownIPV4, err := newIpv4(ownIP)
	if err != nil {
		return err
	}

	// list hosts that know of this file
	goodHosts := s.getGoodHosts(hash, []uint16{}, ownIPV4)
	if len(goodHosts) == 0 {
		return fmt.Errorf("no good hosts found for hash %v", hash)
	}

	// if this download does not already exist, record it in the catalogue
	if extantRecord == nil {
		peer := goodHosts[0]
		fileMeta, err := s.downloadMetaData(hash, peer.Address, peer.Port)
		if err != nil {
			return err
		}
		extantRecord, err = s.cat.RegisterDownload(
			fileMeta.SizeInBytes,
			fileMeta.ChunkCount,
			fileMeta.ChunkSizeInBytes,
			fileMeta.Sha1Hash,
			fileMeta.FileName,
		)
		if err != nil {
			return err
		}
	}

	go func() { // TODO: make interruptiple with a channel
		for !s.cat.FileComplete(hash) {

			// MARK: MASSIVE HACK. Fix this first!
			wantedChunks := s.cat.MissingChunks(hash, 1) // TOOD: something much better than just
			// getting the next wanted chunk. This runs the entire download serially!
			// Find the least-known chunk from each host and download them
			c := wantedChunks[0]
			goodHosts = s.getGoodHosts(hash, []uint16{c.Start, c.End}, ownIPV4)

			// we should really make a slice of host,chunk pairs.
			for _, hostResponse := range goodHosts {
				ip, port := hostResponse.Address, hostResponse.Port
				key := downloadKey{hash: *hash, remoteHost: hostResponse.Address}

				// start downloading from this host if not doing that already
				s.transferLock.Lock()
				if _, ok := s.downloads[key]; !ok {
					fmt.Printf("Getting chunk: %v\n", c.Start)
					s.downloads[key] = struct{}{}
					s.downloadChunk(ip, port, hash, c.Start) // TODO: make concurrent
				}
				s.transferLock.Unlock()
			}
		}
		fmt.Println("Download complete")
	}()

	return nil
}

func (s *Server) downloadMetaData(
	hash *common.Sha1Hash,
	addr [4]byte,
	port uint16,
) (*messages.ListFilesEntry, error) {
	fileMetaList, err := s.ListFilesOnHost(addr, port, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file metadata: %v", err)
	}

	if len(fileMetaList.Files) != 1 || *fileMetaList.Files[0].Sha1Hash != *hash {
		return nil, fmt.Errorf("flu error: selected peer %v told us nothing about %v", addr, hash)
	}
	fileMeta := fileMetaList.Files[0]
	return &fileMeta, nil
}

func (s *Server) downloadChunk(ip [4]byte, port uint16, sha1Hash *common.Sha1Hash, chunk uint16) {
	conn, err := DialPeer(ip, port, sha1Hash, chunk)
	if err != nil {
		panic(err) // TODO: log somewhere and move on...
	}

	start := time.Now()

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
				// empty packet == download complete
				hash := sha1.New()
				hash.Write(conn.buffer)
				finalHash := (&common.Sha1Hash{}).FromSlice(hash.Sum(nil))
				if finalHash.Data != conn.hash.Data {
					// TODO: Handle this a little more gracefully...
					fmt.Println("Sha1 hashes did not match. Download was corrupted. Retrying.")
				}
				err = s.cat.SaveChunk(sha1Hash, chunk, conn.buffer)
				if err != nil {
					panic(err)
				}
				downloadTime := float64(time.Since(start).Seconds())
				speed := 4.0 / downloadTime
				fmt.Printf("Chunk %d complete at %.2f mbps\n", chunk, speed)
				break
			}
		}
		conn.Ack(packet.Offset)
	}

	delete(s.downloads, downloadKey{hash: *sha1Hash, remoteHost: ip})
}

func (s *Server) getGoodHosts(
	hash *common.Sha1Hash,
	chunks []uint16,
	ownIP [4]byte,
) []*messages.DiscoverHostResponse {
	resps := s.DiscoverHosts(hash, make([]uint16, 0))
	result := make([]*messages.DiscoverHostResponse, 0, len(resps))
	for i := 0; i < len(resps); i++ {
		// skip hosts that know nothing about the file we want
		if len(resps[i].Chunks) > 0 && resps[i].Address != ownIP {
			result = append(result, &resps[i])
		}
	}
	return result
}
