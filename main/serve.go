package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	dtypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"

	"github.com/Dataman-Cloud/puller/debug"
	"github.com/Dataman-Cloud/puller/version"
)

type puller struct {
	cfg     *Config // configs
	dc      *client.Client
	startAt time.Time
}

func newPuller(cfg *Config) (*puller, error) {
	p := &puller{
		cfg:     cfg,
		startAt: time.Now(),
	}

	dc, err := client.NewClientWithOpts()
	if err != nil {
		return nil, err
	}
	p.dc = dc

	return p, p.pingDocker()
}

func (p *puller) run() error {
	logrus.Printf("starting puller daemon and serving on %s ...", p.cfg.Listen)
	http.HandleFunc("/ping", p.pong)
	http.HandleFunc("/pull", p.servePulling)
	http.HandleFunc("/debug/dump", p.debugDump)
	http.HandleFunc("/debug/toggle", p.debugToggle)
	return http.ListenAndServe(p.cfg.Listen, nil)
}

func (p *puller) pingDocker() error {
	_, err := p.dc.Ping(context.TODO())
	return err
}

func (p *puller) pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte{'O', 'K'})
}

func (p *puller) servePulling(w http.ResponseWriter, r *http.Request) {
	if err := p.pingDocker(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var req = new(PullRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := req.valid(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")

	resp := p.doPulls(req)
	if len(resp.Failure) > 0 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	bs, _ := json.Marshal(resp)
	w.Write(bs)
}

func (p *puller) debugDump(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]interface{}{
		"configs":  p.cfg,
		"start_at": p.startAt.Format(time.RFC3339),
		"num_fds":  debug.NumFd(),
		"version":  version.FullVersion(),
	})
}

func (p *puller) debugToggle(w http.ResponseWriter, r *http.Request) {
	switch l := logrus.GetLevel(); l {
	case logrus.DebugLevel:
		logrus.SetLevel(logrus.InfoLevel)
	default:
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func (p *puller) doPulls(req *PullRequest) *PullResponse {
	// set concurrency
	var concurrency = req.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	if concurrency > len(req.Images) {
		concurrency = len(req.Images)
	}

	// set retry
	var retry = req.Retry
	if retry <= 0 {
		retry = 3
	}

	var (
		wg     sync.WaitGroup
		tokens = make(chan struct{}, concurrency)

		ret = &PullResponse{
			Success: make([]*ImagePullResponse, 0, 0),
			Failure: make([]*ImagePullResponse, 0, 0),
			StartAt: time.Now(),
		}
		l sync.Mutex // protect obove
	)

	logrus.Printf("pulling %d images by %d concurrency and with maximum %d failure retry ...", len(req.Images), concurrency, retry)

	wg.Add(len(req.Images))
	for _, ipr := range req.Images {
		tokens <- struct{}{} // take one token to start up one worker

		go func(ipr *ImagePullRequest) {
			defer func() {
				<-tokens // release one token
				wg.Done()
			}()

			logrus.Printf("pulling image %s ...", ipr.imageName())
			resp := p.doPull(ipr, retry)

			l.Lock()
			if resp.ErrMsg != "" {
				ret.Failure = append(ret.Failure, resp)
				logrus.Warnf("pulling image %s failed: %v", ipr.imageName(), resp.ErrMsg)
			} else {
				ret.Success = append(ret.Success, resp)
				logrus.Printf("pulled image %s in %s", ipr.imageName(), resp.Cost)
			}
			l.Unlock()

		}(ipr)
	}
	wg.Wait()
	ret.Cost = time.Now().Sub(ret.StartAt).String()

	logrus.Printf("pulling %d images succeed, %d failed, time cost %s", len(ret.Success), len(ret.Failure), ret.Cost)
	return ret
}

func (p *puller) doPull(ipr *ImagePullRequest, maxRetry int) (resp *ImagePullResponse) {
	var (
		startAt = time.Now()
		retry   int
		dec     = new(json.Decoder)
		err     error // the final error, setup later
	)

	resp = &ImagePullResponse{
		Image:   ipr.Image,
		Tag:     ipr.Tag,
		Project: ipr.Project,
	}
	defer func() {
		resp.Cost = time.Now().Sub(startAt).String()
		resp.Retried = retry
		if err != nil {
			resp.ErrMsg = err.Error()
		}
	}()

	// https://docs.docker.com/develop/sdk/examples/#pull-an-image-with-authentication
	var pullOption = dtypes.ImagePullOptions{}
	if regAuthStr := ipr.encodeAuth(); regAuthStr != "" {
		pullOption.RegistryAuth = regAuthStr
		logrus.Debugln("------> Registry Auth:", regAuthStr) // for debug
	}

RETRY:
	retry++

	stream, err := p.dc.ImagePull(context.TODO(), ipr.imageName(), pullOption)
	if err != nil { // the pulling met any early errors (note: the pulling progress not started yet)
		goto END
	}
	defer stream.Close()

	// hanging wait util the pulling progress finished or met any errors
	// io.Copy(ioutil.Discard, stream)
	// decode each frame of half way messages to see if met any errors
	dec = json.NewDecoder(stream)
	for {
		progress := make(map[string]interface{})
		err = dec.Decode(&progress)
		if err != nil {
			if err == io.EOF {
				err = nil // END: pull normally exit
			} else {
				err = fmt.Errorf("decode pulling stream error: %v", err) // maybe peer docker daemon crashed
			}
			goto END
		}

		// for debug
		bs, _ := json.Marshal(progress)
		logrus.Debugln("------> ", string(bs))

		if msg, ok := progress["error"]; ok {
			err = fmt.Errorf("the pulling progress abort with error: %v", msg) // the pulling progress met any halfway error
			goto END
		}
	}

END:
	if err == nil {
		return
	}

	// met any error and retry retry
	if retry < maxRetry {
		logrus.Warnf("[%d] pulling image %s [%v], try again ...", retry, ipr.imageName(), err)
		time.Sleep(time.Millisecond * 500 * time.Duration(retry))
		goto RETRY
	}

	logrus.Warnf("[%d] pulling image %s [%v], no more retry", retry, ipr.imageName(), err)
	return
}
