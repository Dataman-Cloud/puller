package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func atomicWriteFile(path string, data []byte, mode os.FileMode) (err error) {
	f, err := ioutil.TempFile("", "atomic-temp-")
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			f.Close()
			os.Remove(f.Name())
		}
	}()

	n, err := f.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return fmt.Errorf("AtomicWriteFile wrote less than expected")
	}

	if err := f.Sync(); err != nil {
		return err
	}

	f.Close()
	os.Chmod(f.Name(), mode)
	return os.Rename(f.Name(), path) // rename sysCall is atomic under linux
}
