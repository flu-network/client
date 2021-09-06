package catalogue

// IndexRecord describes a file that is 'known' by the flu client. The existence of an IndexRecord
// does not imply that the file exists locally. To find out which chunks of the file are
// downloaded, consult the progressFile. By convention, the progressFile is always named after
// the sha1 of the completely-downloaded file.
type IndexRecord struct {
	FilePath     string
	SizeInBytes  int
	Sha1Hash     [20]byte
	ProgressFile *ProgressFile
}
