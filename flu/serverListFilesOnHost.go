package flu

import (
	"fmt"
	"log"
	"net"
	"path"
	"time"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/common"
	"github.com/flu-network/client/flu/messages"
)

// ListFilesOnHost sends a request for a list of files from a remote host and returns the response.
// The request times out if not fulfilled in a few seconds
func (s *Server) ListFilesOnHost(
	ipv4 [4]byte,
	port uint16,
	hash *common.Sha1Hash,
) (*messages.ListFilesResponse, error) {
	// construct a request and set up UDP harness
	req := messages.ListFilesRequest{RequestID: s.generateRequestID(), Sha1Hash: hash}
	targetAddr := net.UDPAddr{IP: ipv4[:], Port: int(port)}
	conn, err := net.DialUDP("udp", nil, &targetAddr)
	check(err)
	defer conn.Close()

	// setup response harness. Buffered so we don't deadlock
	responseChan := make(chan []byte, 1)
	errorChan := make(chan error, 1)

	// send the request and catch the response
	go func() {
		err = conn.SetReadDeadline(time.Now().Add(time.Second * 2))
		if err != nil {
			errorChan <- err
		} else {
			conn.Write(req.Serialize())
			rawResponseBuffer := make([]byte, 1024)
			_, err = conn.Read(rawResponseBuffer)
			if err != nil {
				errorChan <- err
			} else {
				responseChan <- rawResponseBuffer
			}
		}
	}()

	// parse and return the response
	select {
	case rawResponse := <-responseChan:
		parsedResponse, err := messages.Parse(rawResponse)
		if err != nil {
			return nil, err
		}
		result, ok := parsedResponse.(*messages.ListFilesResponse)
		if !ok {
			return nil, fmt.Errorf("wrong response type received for ListFilesRequest: %v", result)
		}
		return result, nil
	case err = <-errorChan:
		return nil, err
	}
}

func (s *Server) RespondToListFilesOnHost(
	req *messages.ListFilesRequest,
	conn *net.UDPConn,
	returnAddr *net.UDPAddr,
) error {
	var files []catalogue.IndexRecordExport
	var err error

	if req.Sha1Hash.IsBlank() {
		files, err = s.cat.ListFiles()
		if err != nil {
			log.Fatal(err)
		}
	} else {
		file, err := s.cat.Contains(req.Sha1Hash)
		if err != nil {
			log.Fatal(err)
		}
		files = append(files, *file)
	}

	resp := messages.ListFilesResponse{
		RequestID: req.RequestID,
		Files:     make([]messages.ListFilesEntry, len(files)),
	}
	for i, f := range files {
		_, fileName := path.Split(f.FilePath)
		hash := (&common.Sha1Hash{}).FromSlice(f.Sha1Hash.Slice())
		resp.Files[i] = messages.ListFilesEntry{
			SizeInBytes:      uint64(f.SizeInBytes),
			ChunkCount:       uint32(f.Progress.Size()),
			ChunkSizeInBytes: uint32(f.ChunkSize),
			ChunksDownloaded: uint32(f.Progress.Count()),
			Sha1Hash:         hash,
			FileName:         fileName,
		}
	}

	_, err = conn.WriteToUDP(resp.Serialize(), returnAddr)
	return err
}
