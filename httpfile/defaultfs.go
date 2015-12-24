package httpfile

var DefaultFS = &FS{}

// Returns the length of the resource in bytes.
func GetLength(url string) (ret int64, err error) {
	return DefaultFS.GetLength(url)
}
