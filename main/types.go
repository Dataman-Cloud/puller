package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	dtypes "github.com/docker/docker/api/types"

	"github.com/Dataman-Cloud/puller/pkg/validator"
)

// Config is exported
type Config struct {
	Listen       string `json:"listen"`        // must
	DockerSocket string `json:"docker_socket"` // docker unix socket path
}

func (c *Config) valid() error {
	if err := validator.String(c.Listen, 1, 1024, nil); err != nil {
		return fmt.Errorf("listen addr %v", err)
	}
	if c.DockerSocket == "" {
		return errors.New("docker unix sokcet path required")
	}
	return nil
}

// PullRequest is exported
type PullRequest struct {
	Concurrency int                 `json:"concurrency"` // max concurrency nb
	Retry       int                 `json:"retry"`       // max retry nb
	Images      []*ImagePullRequest `json:"images"`      // image list
}

func (r *PullRequest) valid() error {
	if len(r.Images) == 0 {
		return errors.New("empty image list")
	}
	for _, ipr := range r.Images {
		if err := ipr.valid(); err != nil {
			return err
		}
	}
	return nil
}

// ImagePullRequest is exported
type ImagePullRequest struct {
	Image      string             `json:"image"`
	Tag        string             `json:"tag"`
	Project    string             `json:"project"`               // meaningless, just attach with response
	AuthConfig *dtypes.AuthConfig `json:"auth_config,omitempty"` // parameters for authenticating with the docker registry
}

func (r *ImagePullRequest) valid() error {
	if r.Image == "" {
		return errors.New("image required")
	}
	if r.Tag == "" {
		return errors.New("tag required")
	}
	return nil
}

func (r *ImagePullRequest) imageName() string {
	return fmt.Sprintf("%s:%s", r.Image, r.Tag)
}

func (r *ImagePullRequest) withAuth() bool {
	return r.AuthConfig != nil
}

func (r *ImagePullRequest) encodeAuth() string {
	if !r.withAuth() {
		return ""
	}
	bs, _ := json.Marshal(r.AuthConfig)
	return base64.URLEncoding.EncodeToString(bs)
}

// PullResponse is exported
type PullResponse struct {
	Success []*ImagePullResponse `json:"success"` // success list
	Failure []*ImagePullResponse `json:"failure"` // failure list
	StartAt time.Time            `json:"startAt"` // start at time
	Cost    string               `json:"cost"`    // total time cost
}

// ImagePullResponse is exported
type ImagePullResponse struct {
	Image   string `json:"image"`   // same as .Request.Image
	Tag     string `json:"tag"`     // same as .Request.Tag
	Project string `json:"project"` // same as .Request.Project
	Cost    string `json:"cost"`    // time cost
	Retried int    `json:"retried"` // retried nb
	ErrMsg  string `json:"errmsg"`  // error message
}
