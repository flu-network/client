package messages

import "fmt"

// Parse attempts to parse a serialized message into a flu messages.Message
func Parse(data []byte) (msg Message, err error) {
	reader := byteReader{Data: data}
	msgType := reader.readByte()

	switch msgType {
	case discoverHostRequest:
		reqID := reader.readUint16()
		hash := reader.readSha1Hash()
		chunks := reader.readSliceUint16()
		return &DiscoverHostRequest{
			Sha1Hash:  *hash,
			RequestID: reqID,
			Chunks:    chunks,
		}, nil

	case discoverHostResponse:
		reqID := reader.readUint16()
		addr := reader.readBytes(4)
		port := reader.readUint16()
		chunks := reader.readSliceUint16()
		return &DiscoverHostResponse{
			Address:   [4]byte{addr[0], addr[1], addr[2], addr[3]},
			Port:      port,
			RequestID: reqID,
			Chunks:    chunks,
		}, nil

	case listFilesRequest:
		reqID := reader.readUint16()
		hash := reader.readSha1Hash()
		return &ListFilesRequest{
			RequestID: reqID,
			Sha1Hash:  hash,
		}, nil

	case listFilesResponse:
		reqID := reader.readUint16()
		entryCount := reader.readUint16()
		entries := make([]ListFilesEntry, entryCount)
		for i := 0; i < int(entryCount); i++ {
			entries[i] = ListFilesEntry{
				SizeInBytes:      reader.readUint64(),
				ChunkCount:       reader.readUint32(),
				ChunkSizeInBytes: reader.readUint32(),
				ChunksDownloaded: reader.readUint32(),
				Sha1Hash:         reader.readSha1Hash(),
				FileName:         reader.readString256(),
			}
		}
		return &ListFilesResponse{
			RequestID: reqID,
			Files:     entries,
		}, nil

	case openLineRequest:
		sha1Hash := reader.readSha1Hash()
		chunk := reader.readUint16()
		cap := reader.readUint16()
		return &OpenConnectionRequest{
			Sha1Hash:  sha1Hash,
			Chunk:     chunk,
			WindowCap: cap,
		}, nil

	case dataPacketAck:
		offset := reader.readUint32()
		return &DataPacketAck{Offset: offset}, nil

	default:
		return nil, fmt.Errorf("Message of unknown type discarded: %d", msgType)
	}
}
