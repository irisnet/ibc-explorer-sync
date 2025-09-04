# ibc-explorer-sync
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

A daemon that synchronizes on-chain data of Cosmos-SDK based chains into a relational database, with special support for IBC (Inter-Blockchain Communication) data synchronization.

## Description of Branches

### [cosmos-sync](https://github.com/irisnet/ibc-explorer-sync/tree/cosmos-sync) 

- Compatible with most Cosmos ecosystem chains for IBC data synchronization, supporting cosmos-sdk v0.47.5 & cometbft v0.37.2.
- Not suitable for chains with multiple address prefixes.
- Not suitable for chains that already have dedicated branch code.

### [evmos](https://github.com/irisnet/ibc-explorer-sync/tree/evmos)

- A server that synchronizes **Evmos**  chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.

### [multiprefix](https://github.com/irisnet/ibc-explorer-sync/tree/multiprefix)

- A server that synchronizes **LikeCoin** or **Shentu**   chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.
- Chain Code Version Requirements: Requires cosmos-sdk version < v0.47 & tendermint framework.

### [injective](https://github.com/irisnet/ibc-explorer-sync/tree/injective)

- A server that synchronizes **Injective**  chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.

## [cronos](https://github.com/irisnet/ibc-explorer-sync/tree/cronos)

- A server that synchronizes **Cronos**  chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.


## [okchain](https://github.com/irisnet/ibc-explorer-sync/tree/okchain)

- A server that synchronizes **OKXChain**  chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.


## [uptick-stride](https://github.com/irisnet/ibc-explorer-sync/tree/uptick-stride)

- A server that synchronizes **Uptick** or **Stride** chain data into a MongoDB database, with special support for IBC (Inter-Blockchain Communication) data synchronization.
