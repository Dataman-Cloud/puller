package version

import (
	"html/template"
	"io"
	"runtime"
)

// nolint
var (
	version   string = "1.0" // set by build LD_FLAGS
	gitCommit string         // set by build LD_FLAGS
	buildAt   string         // set by build LD_FLAGS
)

// GetVersion is exported
func GetVersion() string {
	return version
}

// GetGitCommit is exported
func GetGitCommit() string {
	return gitCommit
}

// FullVersion is exported
func FullVersion() Version {
	return Version{
		Version:   version,
		GitCommit: gitCommit,
		BuildTime: buildAt,
		GoVersion: runtime.Version(),
		Os:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

var versionTemplate = ` Version:      {{.Version}}
 Git commit:   {{.GitCommit}}
 Go version:   {{.GoVersion}}
 Built:        {{.BuildTime}}
 OS/Arch:      {{.Os}}/{{.Arch}}
`

// Version is exported
type Version struct {
	Version   string `json:"version"`
	GoVersion string `json:"go_version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

// WriteTo is exported
func (v Version) WriteTo(w io.Writer) (int64, error) {
	tmpl, _ := template.New("version").Parse(versionTemplate)
	return -1, tmpl.Execute(w, v) // just make pass govet
}
