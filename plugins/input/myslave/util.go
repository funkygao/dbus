package myslave

import (
	"io"
	"os"
)

// Write file to temp and atomically move when everything else succeeds.
func writeFileAtomic(filename string, data []byte, perm os.FileMode) error {
	f, err := os.OpenFile(filename+".tmp", os.O_CREATE|os.O_RDWR|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	n, err := f.Write(data)
	f.Close()
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	} else {
		err = os.Chmod(f.Name(), perm)
	}
	if err != nil {
		os.Remove(f.Name())
		return err
	}
	return os.Rename(f.Name(), filename)
}
