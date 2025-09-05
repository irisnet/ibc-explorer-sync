package main

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/irisnet/ibc-explorer-sync/config"
	"github.com/irisnet/ibc-explorer-sync/handlers"
	"github.com/irisnet/ibc-explorer-sync/libs/logger"
	"github.com/irisnet/ibc-explorer-sync/libs/pool"
	"github.com/irisnet/ibc-explorer-sync/models"
	"github.com/irisnet/ibc-explorer-sync/tasks"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() / 2)
	c := make(chan os.Signal)

	defer func() {
		logger.Info("System Exit")

		models.Close()

		if err := recover(); err != nil {
			logger.Error("occur error", logger.Any("err", err))
			os.Exit(1)
		}
	}()

	conf, err := config.ReadConfig()
	if err != nil {
		logger.Fatal(err.Error())
	}
	models.Init(conf)
	pool.Init(conf)
	handlers.InitRouter(conf)

	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	tasks.Start(tasks.NewSyncTask(conf))
	<-c
}
