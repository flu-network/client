package flu

import (
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
	ipv4 *[4]byte,
	port uint16,
	hash *common.Sha1Hash,
) *messages.ListFilesResponse {
	// construct a request
	req := messages.ListFilesRequest{
		RequestID: s.generateRequestID(),
		Sha1Hash:  hash,
	}

	// add a response harness for it
	responseChan := s.registerResponseChan(req.RequestID, req.ResponseType())

	// dispatch it
	targetAddr := net.UDPAddr{
		IP:   []byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
		Port: int(port),
	}
	conn, err := net.DialUDP("udp", nil, &targetAddr)
	check(err)
	defer conn.Close()
	conn.Write(req.Serialize())

	// set a timeout and wait for the response
	waitChan := time.After(2 * time.Second)
	select {
	case <-waitChan:
		s.unregisterResponseChan(req.RequestID, req.ResponseType())
		return nil
	case res := <-responseChan:
		parsedResponse, ok := res.(*messages.ListFilesResponse)
		if !ok {
			log.Fatal(ok)
		}
		return parsedResponse
	}
}

func (s *Server) RespondToListFilesOnHost(req *messages.ListFilesRequest) []byte {
	var files []catalogue.IndexRecord
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
		resp.Files[i] = messages.ListFilesEntry{
			SizeInBytes:      uint64(f.SizeInBytes),
			ChunkCount:       uint32(f.ProgressFile.Progress.Size()),
			ChunkSizeInBytes: uint32(f.ChunkSize),
			ChunksDownloaded: uint32(f.ProgressFile.Progress.Count()),
			Sha1Hash:         &f.Sha1Hash,
			FileName:         fileName,
		}
	}

	return resp.Serialize()
}
