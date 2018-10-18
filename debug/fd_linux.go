package debug

import (
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

var (
	sysProc = "/proc" // linux only
)

// NumFd return current process's fd count
func NumFd() int {
	pid := os.Getpid()
	files, _ := ioutil.ReadDir(path.Join(sysProc, strconv.Itoa(pid), "fd"))
	return len(files)
}
