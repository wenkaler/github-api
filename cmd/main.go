package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitlog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/jasonlvhit/gocron"
	"github.com/kelseyhightower/envconfig"
	"github.com/wenkaler/github-api/pkg/github"
)

type configuration struct {
	Project string `envconfig:"git_project"`
	Owner   string `envconfig:"git_owner"`
	Token   string `envconfig:"git_token"`
	Email   string `envconfig:"git_email"`
	Name    string `envconfig:"git_name"`
	Time    string `envconfig:"time"`
}

func main() {

	logger := kitlog.NewLogfmtLogger(kitlog.NewSyncWriter(os.Stderr))
	logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)

	var cfg configuration
	if err := envconfig.Process("", &cfg); err != nil {
		level.Error(logger).Log("msg", "failed to load configuration", "err", err)
		os.Exit(1)
	}

	c, err := github.New(&github.Config{
		Project: cfg.Project,
		Owner:   cfg.Owner,
		Token:   cfg.Token,
	})
	if err != nil {
		level.Error(logger).Log("msg", "failed to create client", "err", err)
		os.Exit(1)
	}
	task := func() {
		date := time.Now().Format("02.01.2006")
		err := c.CreateFile(context.Background(), date, cfg.Name, cfg.Email, date, date)
		if err != nil {
			level.Error(logger).Log("msg", "failed create file", "err", err)
		}
	}
	gocron.Every(1).Days().At(cfg.Time).Do(task)
	cronCh := gocron.Start()
	cl := make(chan os.Signal, 1)
	signal.Notify(cl, syscall.SIGTERM, syscall.SIGINT)
	sig := <-cl
	cronCh <- true
	level.Info(logger).Log("msg", "received signal, exiting", "signal", sig)
	level.Info(logger).Log("msg", "goodbye")
}
