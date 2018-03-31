eth-generator
=============

A ethereum key generator and importer in MySQL/Mariadb database

Compilation
-----------

Make sure `golang` and `git` are installed:

```shell
$ sudo apt-get install golang-go git
```

Then, install `eth-generator` wherever you want on the filesystem:

```shell
$ export GOPATH=${HOME}/go
$ git clone https://github.com/mycroft/eth-generator
$ cd eth-generator
$ go get -d -v
$ go build
$ ./eth-generator -h
Usage of ./eth-generator:
  -db-host string
        DB Hostname (default "localhost")
  -db-name string
        DB Name (mandatory)
  -db-pass string
        DB Pass
  -db-user string
        DB User
  -debug
        Debug mode
  -num int
        Num of keys to generate (default 1)
```

Configuration
-------------

`eth-generator` is using a `config.ini` INI file for configuration:

```ini
[general]
debug = false

[keys]
# Number of keys in pool
num = 10

[db]
# Set true to disable DB
# disabled = true

# DB Configuration
host = 172.17.0.2
name = mysql
user = root
pass = pass
```

Usage
-----

## Create tables (initialization)

```shell
$ ./eth-generator -debug -init
2018/03/31 07:14:29 Key pool size is 3.
2018/03/31 07:14:29 Tables created.
$
```

## Generate new keys, if needed

```shell
$ ./eth-generator -debug
2018/03/31 08:40:56 Key pool size is 3.
2018/03/31 08:40:56 Required to create 3 new keys.
2018/03/31 08:40:56 Priv: 304baaf17517f2311029a294a79af9e192bb88f0daf8beaaea50ca197467b9b2
2018/03/31 08:40:56 Pub:  0x39d5b09767129129f3d4f82871e37e416688d503
2018/03/31 08:40:56 DB Store: 39d5b09767129129f3d4f82871e37e416688d503 304baaf17517f2311029a294a79af9e192bb88f0daf8beaaea50ca197467b9b2
2018/03/31 08:40:56 DB Store: returned id:1
2018/03/31 08:40:56 Priv: cd9b76a312dc78a80c8b638b8bb808c8792d874c0a9ee4537d81be7c7f2479a0
2018/03/31 08:40:56 Pub:  0x8394bcd3397fb7929aaec786f33990a996132c06
2018/03/31 08:40:56 DB Store: 8394bcd3397fb7929aaec786f33990a996132c06 cd9b76a312dc78a80c8b638b8bb808c8792d874c0a9ee4537d81be7c7f2479a0
2018/03/31 08:40:56 DB Store: returned id:2
2018/03/31 08:40:56 Priv: 78beeb545d4597ad7266602a1f0271919d896e55776143297be455e328ab871f
2018/03/31 08:40:56 Pub:  0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9
2018/03/31 08:40:56 DB Store: fbfe44c6f9a060112d61b047fc70ab13904cc1d9 78beeb545d4597ad7266602a1f0271919d896e55776143297be455e328ab871f
2018/03/31 08:40:56 DB Store: returned id:3
$
```

It will store in database:

```sql
MariaDB [mysql]> select id, pub, tx_metadata, tx_value, received, used, completed from ethkeys;
+----+------------------------------------------+-------------+----------+----------+------+-----------+
| id | pub                                      | tx_metadata | tx_value | received | used | completed |
+----+------------------------------------------+-------------+----------+----------+------+-----------+
|  1 | 0190d0bfc8b4fea5a266ac35e6c44651a00bd960 | NULL        | 0        | 0        |    0 |         0 |
|  2 | ce2c74a2e40c363e0ed3a94ac4b53bf4abd20121 | NULL        | 0        | 0        |    0 |         0 |
|  3 | 3f4997d8d45952e7eca74be6f4666edeb0d584d4 | NULL        | 0        | 0        |    0 |         0 |
+----+------------------------------------------+-------------+----------+----------+------+-----------+
3 rows in set (0.00 sec)
```

If no new key is needed, `eth-generator` will promptly quit:

```shell
$ ./eth-generator -debug
2018/03/31 08:41:35 Key pool size is 3.
2018/03/31 08:41:35 No need to insert new records (3 keys in DB).
$
```

## Show key statuses

```shell
$ ./eth-generator -status
2018/03/31 08:41:48 id:1 0x39d5b09767129129f3d4f82871e37e416688d503 used:false waited:0 received:0 
2018/03/31 08:41:48 id:2 0xCE3A0be91053acfd3Eb71de4df4423416e978F50 used:false waited:0 received:0 
2018/03/31 08:41:48 id:3 0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9 used:false waited:0 received:0 
$
```

## Watch for transactions

By default, eth-generator will not watch for transactions if transaction is:
* not used (used = false in database);
* started_ts < NOW() - 24h (it will start to watch only on addresses that were asked to be looked the last 24h, no more, to avoid to query too long the API.);
* completed (completed = true in database);
* received >= tx_value (won't watch any more if balance is bigger than waited value).

Therefore, the upstream app must inform eth-generator to watch for transaction using the following query, updating `used`, `tx_value` and `started_ts` fields:

```sql
MariaDB [mysql]> update ethkeys set used = true, tx_value = 1000000000000000000, started_ts = NOW() where pub = 'CE3A0be91053acfd3Eb71de4df4423416e978F50';
Query OK, 1 row affected (0.01 sec)
Rows matched: 1  Changed: 1  Warnings: 0
```

Status will then show:

```shell
$ ./eth-generator -status
2018/03/31 08:42:14 id:1 0x39d5b09767129129f3d4f82871e37e416688d503 used:false waited:0 received:0 
2018/03/31 08:42:14 id:2 0xCE3A0be91053acfd3Eb71de4df4423416e978F50 used:true waited:1000000000000000000 received:0 started_ts:'2018-03-31 06:42:13'
2018/03/31 08:42:14 id:3 0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9 used:false waited:0 received:0 
$
```

Note that once used, the address is no longer in the pool and calling `eth-generator` without any flag will create a new address:

```shell
$ ./eth-generator -debug
2018/03/31 09:06:15 Key pool size is 3.
2018/03/31 09:06:15 Required to create 1 new keys.
2018/03/31 09:06:15 Priv: afa7bc32a0a4f9459b64decf8895c1e54f1d2682ab80d8c69c28913d49483462
2018/03/31 09:06:15 Pub:  0xe9367d67cb8c6d7ba1e591a9b6a5742f6c466e50
2018/03/31 09:06:15 DB Store: e9367d67cb8c6d7ba1e591a9b6a5742f6c466e50 afa7bc32a0a4f9459b64decf8895c1e54f1d2682ab80d8c69c28913d49483462
2018/03/31 09:06:15 DB Store: returned id:4
$
```

Running the watcher is easy. Just use the `-watch` flag:

```shell
$ ./eth-generator -debug -watch
2018/03/31 08:53:42 Key pool size is 3.
2018/03/31 08:53:42 Watch()
2018/03/31 08:53:42 Looking for key 0xCE3A0be91053acfd3Eb71de4df4423416e978F50
2018/03/31 08:53:42 HTTP: Doing HTTP Query: https://api.etherscan.io/api?module=account&action=balance&address=0xCE3A0be91053acfd3Eb71de4df4423416e978F50&tag=latest
2018/03/31 08:53:42 HTTP: Got result {"status":"1","message":"OK","result":"147004910000000"}
2018/03/31 08:53:42 API: Addr: CE3A0be91053acfd3Eb71de4df4423416e978F50 - Status: OK - Balance: 147004910000000
2018/03/31 08:53:42 API BigInt: 147004910000000
2018/03/31 08:53:42 Current: pub:CE3A0be91053acfd3Eb71de4df4423416e978F50 received:0
2018/03/31 08:53:42 Storing new received value (147004910000000) in database.
2018/03/31 08:53:42 UpdateValue pub:CE3A0be91053acfd3Eb71de4df4423416e978F50 value:147004910000000 completed without error
2018/03/31 08:53:42 Query for TX for 0xCE3A0be91053acfd3Eb71de4df4423416e978F50
2018/03/31 08:53:42 HTTP: Doing HTTP Query: https://api.etherscan.io/api?module=account&action=txlist&address=0xCE3A0be91053acfd3Eb71de4df4423416e978F50&startblock=0&endblock=99999999&sort=asc
2018/03/31 08:53:42 HTTP: Got result [...snapped...]
2018/03/31 08:53:42 DB StoreTXs: 4 records
2018/03/31 08:53:42 DB StoreTXs: Storing tx hash:0xd1584fab28a1671f89f5c1b9ceac3cbe24fbd49aa55525dba1764f82c4150bf4
2018/03/31 08:53:42 DB StoreTXs: returned id:0
2018/03/31 08:53:42 DB StoreTXs: Storing tx hash:0x54b7dc49698c76edffde02a8523c9c7433dced517d7f2f93593680c0a72adf26
2018/03/31 08:53:42 DB StoreTXs: returned id:0
2018/03/31 08:53:42 DB StoreTXs: Storing tx hash:0x0010b4398c6ad3f3d93fc81eeca2d16c37e164ccad53fd79b43825753a612041
2018/03/31 08:53:42 DB StoreTXs: returned id:0
2018/03/31 08:53:42 DB StoreTXs: Storing tx hash:0xc8c5ed5821fe765885b79e6d1d77caef0a7375864bcf73cdca3a009c58ec78cd
2018/03/31 08:53:42 DB StoreTXs: returned id:0
2018/03/31 08:53:42 DB StoreTXs Done.
2018/03/31 08:53:42 Watch() done successfully!
$
```

Status will then show:

```shell
$ ./eth-generator -debug -watch
2018/03/31 08:54:27 id:1 0x39d5b09767129129f3d4f82871e37e416688d503 used:false waited:0 received:0 
2018/03/31 08:54:27 id:2 0xCE3A0be91053acfd3Eb71de4df4423416e978F50 used:true waited:147004910000000 received:147004910000000 started_ts:'2018-03-31 06:42:13'
2018/03/31 08:54:27 id:3 0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9 used:false waited:0 received:0 
$
```

And in DB, transactions will be stored:

```mysql
MariaDB [mysql]> select hash, from_addr, value from ethtx where to_addr = '0xCE3A0be91053acfd3Eb71de4df4423416e978F50';
+--------------------------------------------------------------------+--------------------------------------------+-----------------+
| hash                                                               | from_addr                                  | value           |
+--------------------------------------------------------------------+--------------------------------------------+-----------------+
| 0xd1584fab28a1671f89f5c1b9ceac3cbe24fbd49aa55525dba1764f82c4150bf4 | 0xc7a911ac29ea1e3b1d438f98f8bc053131dcaf52 | 57570000000000  |
| 0x0010b4398c6ad3f3d93fc81eeca2d16c37e164ccad53fd79b43825753a612041 | 0xc7a911ac29ea1e3b1d438f98f8bc053131dcaf52 | 150000000000000 |
+--------------------------------------------------------------------+--------------------------------------------+-----------------+
2 rows in set (0.00 sec)
```

Note that in my example, this address gave out eth to a contract. This is why values are not similar to actual balance.

When balance is reached, the address won't be watched anymore:

```shell
$ ./eth-generator -status
2018/03/31 09:00:39 id:1 0x39d5b09767129129f3d4f82871e37e416688d503 used:false waited:0 received:0 
2018/03/31 09:00:39 id:2 0xCE3A0be91053acfd3Eb71de4df4423416e978F50 used:true waited:147004910000000 received:147004910000000 started_ts:'2018-03-31 07:00:24'
2018/03/31 09:00:39 id:3 0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9 used:false waited:0 received:0 
$ ./eth-generator -debug -watch
2018/03/31 09:02:20 Key pool size is 3.
2018/03/31 09:02:20 Watch()
2018/03/31 09:02:20 No record to look after.
2018/03/31 09:02:20 Watch() done successfully!
$
```

At this moment, the upstream app can mark the address as completed:

```mysql
MariaDB [mysql]> update ethkeys set completed = true where pub = 'CE3A0be91053acfd3Eb71de4df4423416e978F50';
Query OK, 1 row affected (0.01 sec)
Rows matched: 1  Changed: 1  Warnings: 0
```

The `completed` addresses won't be reported in status anymore:

```shell
$ ./eth-generator -debug -status
2018/03/31 09:06:43 Key pool size is 3.
2018/03/31 09:06:43 id:1 0x39d5b09767129129f3d4f82871e37e416688d503 used:false waited:0 received:0 
2018/03/31 09:06:43 id:3 0xfbfe44c6f9a060112d61b047fc70ab13904cc1d9 used:false waited:0 received:0 
2018/03/31 09:06:43 id:4 0xe9367d67cb8c6d7ba1e591a9b6a5742f6c466e50 used:false waited:0 received:0 
$
```

With ethereum (geth)
--------------------

## Importing a private key using geth console

```js
> personal.importRawKey('70d881253b6387b34862bb6a546534a0b29b6bc5c96e69a3d700726eaa25e578', '')
"0x3ec9abeab297df036fa6ae51d5603d7c81345a48"
> web3.fromWei(eth.getBalance("0x3ec9abeab297df036fa6ae51d5603d7c81345a48"), "ether")
0
```


Some other notes
----------------

## Launch a Mariadb docker container

```bash
$ docker run --name mariadb -e MYSQL_ROOT_PASSWORD=pass -d mariadb:latest
510cd8837e82580845d9822f7aa1ba6b678872f037f14917ed193fdecbcd712d
$ docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mariadb
172.17.0.2
```

## DB Schemas:

```mysql
MariaDB [mysql]> desc ethkeys;
+-------------+---------------+------+-----+---------------------+----------------+
| Field       | Type          | Null | Key | Default             | Extra          |
+-------------+---------------+------+-----+---------------------+----------------+
| id          | int(11)       | NO   | PRI | NULL                | auto_increment |
| pub         | char(40)      | NO   |     | NULL                |                |
| priv        | char(64)      | NO   |     | NULL                |                |
| used        | tinyint(1)    | NO   |     | 0                   |                |
| completed   | tinyint(1)    | NO   |     | 0                   |                |
| tx_metadata | text          | YES  |     | NULL                |                |
| tx_value    | decimal(32,0) | YES  |     | 0                   |                |
| received    | decimal(32,0) | YES  |     | 0                   |                |
| started_ts  | timestamp     | NO   |     | 2018-01-01 00:00:00 |                |
+-------------+---------------+------+-----+---------------------+----------------+
9 rows in set (0.00 sec)
```

```mysql
MariaDB [mysql]> desc ethtx;
+---------------------+------------------+------+-----+---------+----------------+
| Field               | Type             | Null | Key | Default | Extra          |
+---------------------+------------------+------+-----+---------+----------------+
| id                  | int(11)          | NO   | PRI | NULL    | auto_increment |
| block_hash          | text             | YES  |     | NULL    |                |
| block_number        | int(11)          | YES  |     | NULL    |                |
| confirmations       | int(11)          | YES  |     | NULL    |                |
| contract_address    | text             | YES  |     | NULL    |                |
| cumulative_gas_used | text             | YES  |     | NULL    |                |
| from_addr           | text             | YES  |     | NULL    |                |
| gas                 | text             | YES  |     | NULL    |                |
| gas_price           | text             | YES  |     | NULL    |                |
| hash                | varchar(66)      | NO   | UNI | NULL    |                |
| timestamp           | int(10) unsigned | YES  |     | NULL    |                |
| transaction_index   | int(11)          | YES  |     | NULL    |                |
| to_addr             | text             | YES  |     | NULL    |                |
| value               | text             | YES  |     | NULL    |                |
| tx_receipt_status   | int(11)          | YES  |     | NULL    |                |
| is_error            | int(11)          | YES  |     | NULL    |                |
+---------------------+------------------+------+-----+---------+----------------+
16 rows in set (0.00 sec)
```
