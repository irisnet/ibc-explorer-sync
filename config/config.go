package config

import (
	"os"

	"github.com/irisnet/ibc-explorer-sync/utils/constant"
	"github.com/spf13/viper"
)

var (
	ConfigFilePath string
)

type (
	Config struct {
		DataBase DataBaseConf `mapstructure:"database"`
		Server   ServerConf   `mapstructure:"server"`
	}

	DataBaseConf struct {
		NodeUri  string `mapstructure:"node_uri"`
		Database string `mapstructure:"database"`
	}

	ServerConf struct {
		NodeUrls                  string `mapstructure:"node_urls"`
		WorkerNumCreateTask       int    `mapstructure:"worker_num_create_task"`
		WorkerNumExecuteTask      int    `mapstructure:"worker_num_execute_task"`
		WorkerMaxSleepTime        int    `mapstructure:"worker_max_sleep_time"`
		BlockNumPerWorkerHandle   int    `mapstructure:"block_num_per_worker_handle"`
		SleepTimeCreateTaskWorker int    `mapstructure:"sleep_time_create_task_worker"`

		MaxConnectionNum   int    `mapstructure:"max_connection_num"`
		InitConnectionNum  int    `mapstructure:"init_connection_num"`
		Bech32AccPrefix    string `mapstructure:"bech32_acc_prefix"`
		ChainId            string `mapstructure:"chain_id"`
		Chain              string `mapstructure:"chain"`
		ChainBlockInterval int    `mapstructure:"chain_block_interval"`
		BehindBlockNum     int    `mapstructure:"behind_block_num"`

		PromethousPort  string `mapstructure:"promethous_port"`
		SupportTypes    string `mapstructure:"support_types"`
		IgnoreIbcHeader bool   `mapstructure:"ignore_ibc_header"`
		UseNodeUrls     bool   `mapstructure:"use_node_urls"`
	}
)

func init() {
	configPath, found := os.LookupEnv(constant.EnvNameConfigFilePath)
	if found {
		ConfigFilePath = configPath
	} else {
		ConfigFilePath = "./config/config.toml"
	}
}

func ReadConfig() (*Config, error) {
	rootViper := viper.New()
	rootViper.SetConfigFile(ConfigFilePath)

	// Read the config file
	if err := rootViper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := rootViper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Override with environment variables if set
	if dbUri := os.Getenv("DB_URI"); dbUri != "" {
		cfg.DataBase.NodeUri = dbUri
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.DataBase.Database = dbName
	}
	if nodeUrls := os.Getenv("NODE_URLS"); nodeUrls != "" {
		cfg.Server.NodeUrls = nodeUrls
	}
	if chainId := os.Getenv("CHAIN_ID"); chainId != "" {
		cfg.Server.ChainId = chainId
	}
	if chain := os.Getenv("CHAIN_NAME"); chain != "" {
		cfg.Server.Chain = chain
	}
	if prefix := os.Getenv("BECH32_PREFIX"); prefix != "" {
		cfg.Server.Bech32AccPrefix = prefix
	}

	// Calculate sleep time of create task goroutine (in seconds)
	sleepTimeCreateTaskWorker := (cfg.Server.BlockNumPerWorkerHandle * cfg.Server.ChainBlockInterval) / 5
	if sleepTimeCreateTaskWorker == 0 {
		sleepTimeCreateTaskWorker = 1
	}
	cfg.Server.SleepTimeCreateTaskWorker = sleepTimeCreateTaskWorker

	return &cfg, nil
}
