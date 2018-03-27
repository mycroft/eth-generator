eth-generator
=============

A ethereum key generator and importer in MySQL/Mariadb database

Compilation
-----------

With only golang installed:

```shell
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

Usage
-----

```shell
./eth-generator -db-name testaroo -db-host 172.17.0.2 -db-user testaroo -db-pass testaroo -num 3 -debug
2018/03/27 19:07:37 Storing using testaroo:testaroo@tcp(172.17.0.2)/testaroo
2018/03/27 19:07:37 Priv: 70d881253b6387b34862bb6a546534a0b29b6bc5c96e69a3d700726eaa25e578
2018/03/27 19:07:37 Pub:  0x3ec9abeab297df036fa6ae51d5603d7c81345a48
2018/03/27 19:07:37 Record: 23
2018/03/27 19:07:37 Priv: 8fb1b22c5ea9c23744abb37e78718fc4ed3b594f1866f11f1620474716f80741
2018/03/27 19:07:37 Pub:  0xa2a6bb89b2717d4c19f89a81997f66bd84a3da26
2018/03/27 19:07:37 Record: 24
2018/03/27 19:07:37 Priv: cbb8edd06abfcaace1ee89071b6d2e8ed20bce6a46d766eeccc3e74d5aa4a6ef
2018/03/27 19:07:37 Pub:  0x6138eae65472973c620a070217558309c001cb1e
2018/03/27 19:07:37 Record: 25
```

Will store in database:

```sql
MariaDB [testaroo]> SELECT * FROM ethkeys WHERE id >= 23;
+----+------------------------------------------+------------------------------------------------------------------+
| id | pub                                      | priv                                                             |
+----+------------------------------------------+------------------------------------------------------------------+
| 23 | 3ec9abeab297df036fa6ae51d5603d7c81345a48 | 70d881253b6387b34862bb6a546534a0b29b6bc5c96e69a3d700726eaa25e578 |
| 24 | a2a6bb89b2717d4c19f89a81997f66bd84a3da26 | 8fb1b22c5ea9c23744abb37e78718fc4ed3b594f1866f11f1620474716f80741 |
| 25 | 6138eae65472973c620a070217558309c001cb1e | cbb8edd06abfcaace1ee89071b6d2e8ed20bce6a46d766eeccc3e74d5aa4a6ef |
+----+------------------------------------------+------------------------------------------------------------------+
3 rows in set (0.00 sec)
```

To generate keys without storing, do not specify any `-db-name` flag:

```shell
$ ./eth-generator -debug -num 2
2018/03/27 19:12:36 Priv: 15a5c239de918a8431cf39c501d988ef6d39fb83f375a04f984404f65a5a444a
2018/03/27 19:12:36 Pub:  0x7a8851c362026349709fd1da3dcdb81cdab1eea9
2018/03/27 19:12:36 Priv: 0c3ce06e0a471790ab42229537715de69b594ca5f6a7f294c89415ce4e7d81e0
2018/03/27 19:12:36 Pub:  0x3370920532b7d3b214c4f76d73e9cadcf9831459
```

With ethereum
-------------

### Using geth (shell)


### Using geth console

```js
> personal.importRawKey('70d881253b6387b34862bb6a546534a0b29b6bc5c96e69a3d700726eaa25e578', '')
"0x3ec9abeab297df036fa6ae51d5603d7c81345a48"
> web3.fromWei(eth.getBalance("0x3ec9abeab297df036fa6ae51d5603d7c81345a48"), "ether")
0
```