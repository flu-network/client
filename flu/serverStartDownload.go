package flu

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/flu-network/client/common"
	"github.com/flu-network/client/common/bitset"
	"github.com/flu-network/client/flu/messages"
)

// RegisterDownload creates a progressfile for the specified file, adds it to the catalogue, and begins
// the download. A name for the file is chosen arbitrarily from one of the hosts who have that file.
func (s *Server) RegisterDownload(hash *common.Sha1Hash) error {
	extantRecord, _ := s.cat.Contains(hash)

	if extantRecord == nil {
		// list hosts that have info on this file
		hostResponses := s.DiscoverHosts(hash, make([]uint16, 0, 0))

		// from each host, try to get the metadata associated with this file
		var entry *messages.ListFilesEntry
		for _, resp := range hostResponses {
			if len(resp.Chunks) != 0 {
				detailFiles := s.ListFilesOnHost(&resp.Address, resp.Port, hash)
				if len(detailFiles.Files) != 0 {
					entry = &detailFiles.Files[0]
					break
				}
			}
		}
		if entry == nil {
			return fmt.Errorf("No online hosts know about %s", hash.String())
		}

		// register the download in the catalogue, so its persisted
		err := s.cat.RegisterDownload(
			entry.SizeInBytes,
			entry.ChunkCount,
			entry.ChunkSizeInBytes,
			entry.Sha1Hash,
			entry.FileName,
		)
		if err != nil {
			return err
		}

	}

	// add it to the in-memory downloads happening right now
	dl, ok := s.downloads[*hash]
	if !ok {
		dl = newDownload(hash)
		s.downloads[*hash] = dl
		go s.startDownload(hash)
	}

	return nil
}

func newDownload(hash *common.Sha1Hash) *download {
	result := &download{
		downloadLines: make(map[uint16]*downloadLine),
		tokens:        make(chan struct{}, 5),
	}
	return result
}

type download struct {
	downloadLines map[uint16]*downloadLine
	tokens        chan struct{}
}

type hostChunk struct {
	host  [4]byte
	port  uint16
	chunk uint16
}

func (s *Server) getDownloadTargets(hash *common.Sha1Hash) []hostChunk {
	ir := s.cat.Get(hash)
	unfilledRanges := ir.ProgressFile.Progress.UnfilledRanges()
	hosts := s.DiscoverHosts(hash, unfilledRanges)

	for _, host := range hosts {
		// TODO: get the k ranges with the lowest overlap
		if len(host.Chunks) > 0 {
			return []hostChunk{{
				host:  host.Address,
				port:  host.Port,
				chunk: host.Chunks[0],
			}}
		}
	}

	return make([]hostChunk, 0, 0)
}

func (s *Server) startDownload(hash *common.Sha1Hash) {
	ir := s.cat.Get(hash)    // index record
	dl := s.downloads[*hash] // download

	for !ir.ProgressFile.Progress.Full() {
		// 	1. get overlap between chunks needed and chunks available
		hostChunks := s.getDownloadTargets(hash) // TODO: exclude any current downloads

		// 	2. if no overlap
		if len(hostChunks) == 0 {
			//  sleep a few seconds (optional, backing off exponentially)
			time.Sleep(time.Millisecond * 1000)
		} else {
			for _, hostedChunk := range hostChunks {
				dl.tokens <- struct{}{} // acquire a token
				line := &downloadLine{
					buffer:     make([]byte, ir.ChunkSize),
					cancelChan: make(chan struct{}),
					waitChan:   make(chan time.Time),
					recvChan:   make(chan []byte),
					bitmap:     bitset.NewBitset(ir.ChunkSize),
				}
				dl.downloadLines[hostedChunk.chunk] = line
				go dl.start(&hostedChunk, line) // releases the token when done
			}
		}
	}
}

func (dl download) start(chunk *hostChunk, line *downloadLine) {
	// CLIENT SHOULD WRITE SOMETHING LIKE THIS:
	fluRecv, chunkSize, err := DialPeer(ip, port, sha1Hash, chunk)
	// fluRecv should track internal state with a bitset.
	// it should tick bits off as they arrive.
	// hasmore should return false when all bits are ticked or the connection times out.
	// contains running goroutines internally that manage state.

	result := make([]byte, chunkSize) // could be 4mb

	for fluRecv.HasMore() { // returns false when done or errored
		data, offset := fluRecv.Read() // blocks execution
		copy(result[offset:offset+len(data)], data)
		// TODO: could lower the memory footprint here by inserting into a skiplist of linked blocks
	}

	fluRecv.close() // sends shutdown signal to server

	// Flush result to disk if fluRecv.done() is true

	// SERVER SHOULD WRITE SOMETHING LIKE THIS:
	// in handleOpenLineRequest:
	fileReader := newFileReader(sha1Hash, chunkNumber)
	// readerInterface wraps a filedescriptor that yields bytes
	fluSender := newFluSender(ip, port, fileReader, doneChan)
	// contains funning goroutines that manage state and send data.
	// calls donechan when client closes connection or times out.
	fluSender.close() // should be called after doneChan, or to interrupt flow and abandon the transfer

	// CLIENT IMPLEMENTATION COULD LOOK LIKE THIS:
	conn, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: chunk.host[:], Port: int(chunk.port)})
	if err != nil {
		log.Fatal(err) // TODO: probably should fail silently here
	}
	defer conn.Close()

	go func() {
		for {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer) // different read to that in main.go. Check if this works?
			if err != nil {
				// ideally conn.close should trigger this and shut the read down. Try and see.
				log.Fatal(err)
			}
			line.recvChan <- buffer[:n]
		}
	}()

	line.kickstart() // send a request for info and get the ball rolling

	for !line.bitmap.Full() {
		select {
		case <-line.cancelChan:
			return
		case <-line.waitChan:
			line.kickstart()
		case data := <-line.recvChan:
			// parse data
			// ack it (updating the bandwidth)
			// save it into the buffer in the right place
		}
	}

	// flush buffer to disk
	<-dl.tokens // release the token
}

type downloadLine struct {
	buffer     []byte
	cancelChan chan struct{}
	waitChan   chan time.Time
	recvChan   chan []byte
	bitmap     *bitset.Bitset
}

func (line *downloadLine) kickstart() {
	req := messages.OpenLineRequest{}
	l.conn.Write( /*data request*/ )
}
