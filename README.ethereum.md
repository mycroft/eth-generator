Ethereum single node testing
============================

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

## Setup a new chain

To setup a new chain for private testing, you'll need a `genesis.json` file, like this one; As we are going to work in `/tmp`, I suggest to put the following contents into `/tmp/genesis.json`:

```json
{
    "config": {
        "chainId": 15,
        "homesteadBlock": 0,
        "eip155Block": 0,
        "eip158Block": 0
    },
    "difficulty": "0x200",
    "gasLimit": "0x8000000",
    "alloc": {
        "94868f6713eebb451d8af85523e7c7137a1c3284": { "balance": "100000000000000000000" }
    }
}
```

Once you wrote `genesis.json`, use `geth` to create your own blockchain (we are going to use an alias for `geth` pointing to `/tmp/go-ethereum/build/bin/geth`):

```shell
$ cd /tmp
$ alias geth=/tmp/go-ethereum/build/bin/geth
$ mkdir test-chain
$ geth --datadir ./test-chain init genesis.json
INFO [03-27|19:29:58] Maximum peer count                       ETH=25 LES=0 total=25
INFO [03-27|19:29:58] Allocated cache and file handles         database=/tmp/test-chain/geth/chaindata cache=16 handles=16
INFO [03-27|19:29:58] Writing custom genesis block 
INFO [03-27|19:29:58] Persisted trie from memory database      nodes=1 size=201.00B time=14.104µs gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B
INFO [03-27|19:29:58] Successfully wrote genesis state         database=chaindata                      hash=72d4f9…36f0d3
INFO [03-27|19:29:58] Allocated cache and file handles         database=/tmp/test-chain/geth/lightchaindata cache=16 handles=16
INFO [03-27|19:29:58] Writing custom genesis block 
INFO [03-27|19:29:58] Persisted trie from memory database      nodes=1 size=201.00B time=9.073µs  gcnodes=0 gcsize=0.00B gctime=0s livenodes=1 livesize=0.00B
INFO [03-27|19:29:58] Successfully wrote genesis state         database=lightchaindata                      hash=72d4f9…36f0d3
```

Your chain is ready. Next step: Launch the `geth`'sdaemon with `console`:

```shell
$ geth --datadir ./test-chain init genesis.json
INFO [03-27|19:30:43] Maximum peer count                       ETH=25 LES=0 total=25
INFO [03-27|19:30:43] Starting peer-to-peer node               instance=Geth/v1.8.4-unstable-80449719/linux-amd64/go1.10
[...]
>
```

Create yourself a personal account then create some coins (this will be your customer):

```js
> personal.newAccount()
Passphrase: 
Repeat passphrase: 
"0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c"
> miner.setEtherbase("0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c")
> miner.start(1)
INFO [03-27|19:33:01] Updated mining threads                   threads=1
INFO [03-27|19:33:01] Starting mining operation 
INFO [03-27|19:33:01] Commit new mining work                   number=1 txs=0 uncles=0 elapsed=389.313µs
INFO [03-27|19:33:02] Successfully sealed new block            number=1 hash=556965…eeb9ca
INFO [03-27|19:33:02] 🔨 mined potential block                  number=1 hash=556965…eeb9ca
> miner.stop()
```

Okay, we mined a block. Let's check balance:

```js
> web3.fromWei(eth.getBalance("0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c"), "ether")
15
```

Great. Let's send some of this money to a new created address:

```shell
$ ./eth-generator -debug -num 1
2018/03/27 19:34:49 Priv: 5086fb5191c7215d024b89e88628b2be2513ab417a7dfe5649e9ef3cb6f22526
2018/03/27 19:34:49 Pub:  0x76e669fd985997bcf7d31c4ae639be2081f428ea
```

In geth console:

```js
> var sender = "0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c";
undefined
> var receiver = "0x76e669fd985997bcf7d31c4ae639be2081f428ea";
undefined
> var amount = web3.toWei(10, "ether");
undefined
> web3.personal.unlockAccount(sender, '', 15000);
true
> eth.sendTransaction({from:sender, to:receiver, value: amount});
INFO [03-27|19:38:26] Submitted transaction                    fullhash=0x3c7eca7d19cf3830779ff8c0de5ddd0d11015a8e388c4b950f747074e54b0bab recipient=0x76e669fd985997bCF7d31C4ae639BE2081F428eA
"0x3c7eca7d19cf3830779ff8c0de5ddd0d11015a8e388c4b950f747074e54b0bab"
```

At this moment, as there are no miner, the transaction is in pending state waiting to be mined and balance is not spent yet. Mine a little more and you'll reciever those 10 coins:

```js
> web3.fromWei(eth.getBalance("0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c"), "ether")
15
> miner.start()
INFO [03-27|19:39:34] Updated mining threads                   threads=0
INFO [03-27|19:39:34] Transaction pool price threshold updated price=18000000000
INFO [03-27|19:39:34] Starting mining operation 
INFO [03-27|19:39:34] Commit new mining work                   number=4 txs=1 uncles=0 elapsed=638.944µs
INFO [03-27|19:39:34] Successfully sealed new block            number=4 hash=768557…acd4f8
INFO [03-27|19:39:34] 🔨 mined potential block                  number=4 hash=768557…acd4f8
> web3.fromWei(eth.getBalance(receiver), "ether")
10
```

Our receiver (marketplace) received the money. But as we didn't imported the private key yet, we can't use it:

```js
> var sender=receiver;
undefined
> var receiver="0x94e15d9d909bbe88e8c0959e3baa211cea3ac68c";
undefined
> sender
"0x76e669fd985997bcf7d31c4ae639be2081f428ea"
> eth.sendTransaction({from:sender, to:receiver, value: amount});
Error: unknown account
```

We have to import the private key of `0x76e669fd985997bcf7d31c4ae639be2081f428ea` (created by eth-generator). We do this using `personal.importRawKey`:

```js
> personal.importRawKey('5086fb5191c7215d024b89e88628b2be2513ab417a7dfe5649e9ef3cb6f22526', '');
"0x76e669fd985997bcf7d31c4ae639be2081f428ea"
```

We can now send our money (don't forget to unlock the account using password (mine: ''), and to mine at least one block again):

```js
> web3.personal.unlockAccount(sender, '', 15000);
true
> var amount = web3.toWei(9.5, "ether");
undefined
> eth.sendTransaction({from:sender, to:receiver, value: amount});
INFO [03-27|19:46:29] Submitted transaction                    fullhash=0x227a9a97e8c93af64c21bd9c0aede6570bd70e9ef22ca31e473edc4bdc07652d recipient=0x94e15D9d909BbE88e8c0959e3baa211CEa3ac68C
"0x227a9a97e8c93af64c21bd9c0aede6570bd70e9ef22ca31e473edc4bdc07652d"
> miner.start(1);
[...]
> miner.stop();
```

And check our final balances:

```js
> web3.fromWei(eth.getBalance(receiver), "ether")
54.500378
> web3.fromWei(eth.getBalance(sender), "ether")
0.499622
> 
```

That's it! :)