# cosmos-sync
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A server that synchronizes **OKXChain** chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.

## Build And Run

- Build: `make all`
- Run: `make run`
- Cross compilation: `make build-linux`

## Configuration

### Environment Variables

The following environment variables can be used to configure the application:

- `CONFIG_FILE_PATH`: Path to the configuration file (optional, default: `./config/config.toml`)
- `DB_URI`: MongoDB connection URI (overrides config file if set)
- `DB_NAME`: Database name (overrides config file if set)
- `NODE_URLS`: Comma-separated list of RPC node URLs (overrides config file if set)
- `CHAIN_ID`: Chain ID (overrides config file if set)
- `CHAIN_NAME`: Chain name (overrides config file if set)
- `BECH32_PREFIX`: Bech32 address prefix (overrides config file if set)

### Configuration File

The application uses a TOML configuration file. A template is provided at `config/config.toml.example`.
Copy this file to `config/config.toml` and update the values according to your environment.

Example configuration structure:

```toml
[database]
node_uri = "mongodb://<username>:<password>@<host>:<port>/?connect=direct&authSource=<database>"
database = "<database_name>"

[server]
node_urls = "<rpc_node_url>"
worker_num_create_task = 1
worker_num_execute_task = 30
worker_max_sleep_time = 120
block_num_per_worker_handle = 100
max_connection_num = 100
init_connection_num = 5
bech32_acc_prefix = "<chain_prefix>"
chain_id = "<chain_id>"
chain = "<chain_name>"
chain_block_interval = 5
behind_block_num = 0
promethous_port = 9090
support_types = "transfer,recv_packet,timeout_packet,acknowledge_packet,update_client,channel_open_confirm"
ignore_ibc_header = true
use_node_urls = true
```

### Configuration Parameters

#### Database Configuration
- `node_uri`: MongoDB connection URI (required)
- `database`: Database name (required)

#### Server Configuration
- `node_urls`: RPC node URLs (required)
- `worker_num_create_task`: Number of task creation workers (default: 1)
- `worker_num_execute_task`: Number of task execution workers (default: 30)
- `worker_max_sleep_time`: Maximum worker sleep time in seconds (default: 120)
- `block_num_per_worker_handle`: Number of blocks per sync task (default: 100)
- `max_connection_num`: Maximum number of connections in the pool (default: 100)
- `init_connection_num`: Initial number of connections in the pool (default: 5)
- `bech32_acc_prefix`: Chain-specific Bech32 address prefix
- `chain_id`: Chain ID
- `chain`: Chain name for collection naming
- `chain_block_interval`: Block interval in seconds (default: 5)
- `behind_block_num`: Number of blocks to wait when retrying failed tasks (default: 0)
- `promethous_port`: Prometheus metrics port (default: 9090)
- `support_types`: Comma-separated list of supported message types
- `ignore_ibc_header`: Whether to ignore IBC header info (default: false)
- `use_node_urls`: Use configured node URLs (true) or fetch from GitHub (false)

### Initial Setup

To start synchronization from a specific block height:

1. Stop the cosmos-sync service
2. Create a sync task in MongoDB:

```javascript
db.sync_task.insert({
    'start_height': NumberLong(<start_block_height>),
    'end_height': NumberLong(0),
    'current_height': NumberLong(0),
    'status': 'unhandled',
    'worker_id': '',
    'worker_logs': [],
    'last_update_time': NumberLong(0)
})
```

Replace `<start_block_height>` with the desired starting block height.