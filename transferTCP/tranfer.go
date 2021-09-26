package transfertcp

import (
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/flu-network/client/catalogue"
	"github.com/flu-network/client/common"
)

// bufferSize is the default buffer size to use for a tcp transfer
const bufferSize = 1024

// SendFile just ships the file over TCP to the client using the most naive method possible
func SendFile(connection net.Conn, cat *catalogue.Cat) error {
	defer connection.Close()

	hash := common.Sha1Hash{}
	byteCount, err := connection.Read(hash.Slice())
	if err != nil {
		return err
	}

	if byteCount != 20 {
		check(fmt.Errorf("Expected 20 byte hash but received %d", byteCount))
	}

	rec, err := cat.Contains(&hash)
	if err != nil {
		return err
	}

	if rec.ProgressFile.Progress.Full() == false {
		return fmt.Errorf("Cannot dispatch file %s. Integrity < 100%%", rec.FilePath)
	}

	fmt.Printf("Client requesting %s\n", rec.FilePath)

	file, err := os.Open(rec.FilePath)
	if err != nil {
		return err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, bufferSize)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}

	fmt.Printf("Sent %s\n", rec.FilePath)
	return nil
}

// fillString is a horrible hack used to serialize stuff to send over the wire
func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}

// GetFile just accepts whatever the server sends and saves it to ~/Downloads.
func GetFile(hash *common.Sha1Hash) {
	connection, err := net.Dial("tcp", "localhost:17969")
	if err != nil {
		panic(err)
	}
	defer connection.Close()
	connection.Write(hash.Slice())

	fmt.Println("Connected to server, start receiving the file name and file size")
	bufferFileName := make([]byte, 64)
	bufferFileSize := make([]byte, 10)

	connection.Read(bufferFileSize)
	fileSize, err := strconv.ParseInt(strings.Trim(string(bufferFileSize), ":"), 10, 64)
	check(err)

	connection.Read(bufferFileName)
	fileName := strings.Trim(string(bufferFileName), ":")

	fullName := path.Join("/tmp/", fileName)
	fmt.Println(fullName)

	newFile, err := os.Create(fullName)
	check(err)
	defer newFile.Close()
	var receivedBytes int64

	for {
		if (fileSize - receivedBytes) < bufferSize {
			io.CopyN(newFile, connection, (fileSize - receivedBytes))
			connection.Read(make([]byte, (receivedBytes+bufferSize)-fileSize))
			break
		}
		io.CopyN(newFile, connection, bufferSize)
		receivedBytes += bufferSize
	}
	fmt.Println("Received file")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
