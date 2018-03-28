Ethereum testnet
================

## Introduction

There are multiple ethereum testnet networks:

* Ropsten (default, used when using `--testnet`; best for reproduce the current ethereum environment);
* Kovan;
* Rinkeby.

Those 3 chains are already big, and unlike a private network, you have no control over them.

## Compilation

First, make sure `golang` is installed:

```shell
$ sudo apt-get install golang-go
```

First, get `go-ethereum` & build it. You'll need only the `geth` binary:

```shell
$ cd /tmp
$ git clone https://github.com/ethereum/go-ethereum.git
[...]
$ cd go-ethereum
$ make all
[...]
$ ls -l build/bin
total 270512
[...]
-rwxrwxr-x. 1 mycroft mycroft 37385424 Mar 27 19:25 geth
[...]
```

## Connect to testnet

Just launch `geth` with the `--testnet` flag:

```
$ cd /tmp
$ alias geth=/tmp/go-ethereum/build/bin/geth
$ mkdir testnet
$ geth --testnet --datadir ./testnet
INFO [03-28|06:45:03] Maximum peer count                       ETH=25 LES=0 total=25
INFO [03-28|06:45:03] Starting peer-to-peer node               instance=Geth/v1.8.3-unstable-b1917ac9/linux-amd64/go1.10
INFO [03-28|06:45:03] Allocated cache and file handles         database=/home/mycroft/dev/eth-generator/my-ethereum/testnet/geth/chaindata cache=768 handles=512
[...]
INFO [03-28|06:45:29] Block synchronisation started 
INFO [03-28|06:45:29] Imported new block headers               count=0   elapsed=2.351ms   number=26217 hash=30feba…9cd256 ignored=192
INFO [03-28|06:45:29] Imported new state entries               count=384 elapsed=4.205µs   processed=95667 pending=14353 retry=0 duplicate=0 unexpected=0
INFO [03-28|06:45:29] Imported new block receipts              count=187 elapsed=31.223ms  number=26025 hash=e46322…bbf8ae size=507.61kB ignored=0
[...]
```

Block synchronization will be slow (the `Ropsten` chain is >= 12 GB as for today - March 2018).
 