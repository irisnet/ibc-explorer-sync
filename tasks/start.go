package tasks

import "github.com/irisnet/ibc-explorer-sync/monitor"

func Start(synctask SyncTask) {
	go synctask.StartCreateTask()
	go synctask.StartExecuteTask()
	go monitor.Start()
}
