package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/urfave/cli"

	_ "github.com/Dataman-Cloud/puller/debug"
	"github.com/Dataman-Cloud/puller/version"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	globalFlags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:   "log-file",
			Usage:  "The log file path",
			EnvVar: "LOG_FILE",
		},
	}
)

var (
	serveFlags = []cli.Flag{
		cli.StringFlag{
			Name:   "listen",
			Usage:  "The address that API service listens on, eg: :80",
			EnvVar: "LISTEN_ADDR",
			Value:  ":9006",
		},
		cli.StringFlag{
			Name:   "docker-socket",
			Usage:  "The docker unix socket file path",
			EnvVar: "DOCKER_SOCKET",
			Value:  "/var/run/docker.sock",
		},
	}
)

func main() {
	app := cli.NewApp()
	app.Name = "puller"
	app.Author = "GuangZheng Zhang"
	app.Email = "zhang.elinks@gmail.com"
	app.Version = version.GetVersion()
	if gitCommit := version.GetGitCommit(); gitCommit != "" {
		app.Version += "-" + gitCommit
	}

	app.Flags = globalFlags

	app.Before = func(c *cli.Context) error {
		var (
			debug   = c.Bool("debug")
			logFile = c.String("log-file")
		)

		logrus.SetLevel(logrus.InfoLevel)
		if debug {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if logFile == "" {
			logrus.SetOutput(os.Stdout)
		} else {
			fd, err := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
			if err != nil {
				return err
			}
			logrus.SetOutput(fd)
		}
		return nil
	}

	app.Commands = []cli.Command{
		versionCommand(),
		runCommand(),
	}

	app.RunAndExitOnError()
}

func versionCommand() cli.Command {
	return cli.Command{
		Name:  "version",
		Usage: "print version",
		Action: func(c *cli.Context) {
			version.FullVersion().WriteTo(os.Stdout)
		},
	}
}

func runCommand() cli.Command {
	return cli.Command{
		Name:   "serve",
		Usage:  "start the puller daemon",
		Flags:  serveFlags,
		Action: runPuller,
	}
}

func runPuller(c *cli.Context) error {
	cfg, err := newPullerConfig(c)
	if err != nil {
		return err
	}

	puller, err := newPuller(cfg)
	if err != nil {
		return err
	}

	return puller.run()
}

func newPullerConfig(c *cli.Context) (*Config, error) {
	var (
		listen       = c.String("listen")
		dockerSocket = c.String("docker-socket")
	)

	cfg := &Config{
		Listen:       listen,
		DockerSocket: dockerSocket,
	}

	return cfg, cfg.valid()
}
