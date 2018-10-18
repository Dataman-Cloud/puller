package debug

import (
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

var (
	fcpu = filepath.Join(os.TempDir(), "puller-cpu.pprof")
	fmem = filepath.Join(os.TempDir(), "puller-mem.pprof")
)

func init() {
	if env := os.Getenv("PROFILE"); env != "" {
		log.Warnln("running with profile enabled, stop profiling by SIGUSR2")
		prof := newProfile()
		if err := prof.start(); err != nil {
			log.Fatalf("could not profile: %v", err)
		}
	}
}

type profile struct {
	fdcpu           *os.File
	fdmem           *os.File
	previsouMemRate int
}

func newProfile() *profile {
	var (
		prof = new(profile)
	)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGUSR2)
		<-c

		log.Println("profile: caught interrupt, stopping profiles")
		prof.stop()
		os.Exit(0)
	}()

	return prof
}

func (p *profile) start() error {
	fd, err := os.Create(fcpu)
	if err != nil {
		return err
	}
	p.fdcpu = fd

	fd, err = os.Create(fmem)
	if err != nil {
		p.fdcpu.Close()
		return err
	}
	p.fdmem = fd

	pprof.StartCPUProfile(p.fdcpu)

	p.previsouMemRate = runtime.MemProfileRate
	runtime.MemProfileRate = 4096

	return nil
}

func (p *profile) stop() {
	pprof.StopCPUProfile()
	if p.fdcpu != nil {
		p.fdcpu.Close()
	}

	pprof.Lookup("heap").WriteTo(p.fdmem, 0)
	p.fdmem.Close()
	runtime.MemProfileRate = p.previsouMemRate
}
