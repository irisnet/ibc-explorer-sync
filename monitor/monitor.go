package monitor

import (
	"time"

	"github.com/irisnet/ibc-explorer-sync/libs/logger"
	"github.com/irisnet/ibc-explorer-sync/libs/pool"
	"github.com/irisnet/ibc-explorer-sync/models"
	"github.com/irisnet/ibc-explorer-sync/resource"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qiniu/qmgo"
)

const (
	NodeStatusNotReachable = 0
	NodeStatusSyncing      = 1
	NodeStatusCatchingUp   = 2

	SyncTaskFollowing  = 1
	SyncTaskCatchingUp = 0
)

type clientNode struct {
	nodeStatus          prometheus.Gauge
	nodeHeight          prometheus.Gauge
	dbHeight            prometheus.Gauge
	nodeTimeGap         prometheus.Gauge
	syncWorkWay         prometheus.Gauge
	syncCatchingTaskNum prometheus.Gauge
}

func NewMetricNode() clientNode {
	nodeHeightMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "status",
		Name:      "node_height",
		Help:      "full node latest block height",
	})
	dbHeightMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "status",
		Name:      "db_height",
		Help:      "sync system database max block height",
	})
	nodeStatusMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "status",
		Name:      "node_status",
		Help:      "full node status(0:NotReachable,1:Syncing,2:CatchingUp)",
	})
	nodeTimeGapMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "status",
		Name:      "node_seconds_gap",
		Help:      "the seconds gap between node block time with sync db block time",
	})
	syncWorkwayMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "",
		Name:      "task_working_status",
		Help:      "sync task working status(0:CatchingUp 1:Following)",
	})
	syncCatchingTaskNumMetric := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "sync",
		Subsystem: "",
		Name:      "task_catching_cnt",
		Help:      "count of sync catchUping task",
	})
	prometheus.MustRegister(nodeHeightMetric)
	prometheus.MustRegister(dbHeightMetric)
	prometheus.MustRegister(nodeStatusMetric)
	prometheus.MustRegister(nodeTimeGapMetric)
	prometheus.MustRegister(syncWorkwayMetric)
	prometheus.MustRegister(syncCatchingTaskNumMetric)
	return clientNode{
		nodeStatus:          nodeStatusMetric,
		nodeHeight:          nodeHeightMetric,
		dbHeight:            dbHeightMetric,
		nodeTimeGap:         nodeTimeGapMetric,
		syncWorkWay:         syncWorkwayMetric,
		syncCatchingTaskNum: syncCatchingTaskNumMetric,
	}
}

func (node *clientNode) Report() {
	for {
		t := time.NewTimer(time.Duration(10) * time.Second)
		select {
		case <-t.C:
			node.nodeStatusReport()
			node.syncCatchUpingReport()
		}
	}
}
func (node *clientNode) nodeStatusReport() {

	nodeurl, _ := resource.GetValidNodeUrl()
	if len(nodeurl) == 0 {
		nodes := pool.PoolValidNodes()
		if len(nodes) > 0 {
			resource.ReloadRpcResourceMap(nodes)
			node.nodeStatus.Set(float64(NodeStatusSyncing))
		} else {
			node.nodeStatus.Set(float64(NodeStatusNotReachable))
		}
	} else {
		node.nodeStatus.Set(float64(NodeStatusSyncing))
	}

	block, err := new(models.Block).GetMaxBlockHeight()
	if err != nil && err != qmgo.ErrNoSuchDocuments {
		logger.Error("query block exception", logger.String("error", err.Error()))
	}
	node.dbHeight.Set(float64(block.Height))

	follow, err := new(models.SyncTask).QueryValidFollowTasks()
	if err != nil {
		logger.Error("query valid follow task exception", logger.String("error", err.Error()))
		return
	}
	if follow && block.Time > 0 {
		timeGap := time.Now().Unix() - block.Time
		node.nodeTimeGap.Set(float64(timeGap))
	}

	if follow {
		node.syncWorkWay.Set(float64(SyncTaskFollowing))
	} else {
		node.syncWorkWay.Set(float64(SyncTaskCatchingUp))
	}
	return
}

func (node *clientNode) syncCatchUpingReport() {
	catchUpTasksNum, err := new(models.SyncTask).QueryCatchUpingTasksNum()
	if err != nil {
		logger.Error("query task exception", logger.String("error", err.Error()))
	}
	node.syncCatchingTaskNum.Set(float64(catchUpTasksNum))
}

func Start() {
	// start monitor
	node := NewMetricNode()
	node.Report()
}
