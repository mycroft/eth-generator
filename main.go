package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"

	"gopkg.in/ini.v1"

	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

type BalanceResp struct {
	Message string
	Result  string
	Status  string
}

type TX struct {
	BlockHash         string `json:"blockHash"`
	BlockNumber       int    `json:"blockNumber,string"`
	Confirmations     int    `json:"confirmations,string"`
	ContractAddress   string `json:"contractAddress"`
	CumulativeGasUsed string `json:"cumulativeGasUsed"`
	From              string `json:"from"`
	Gas               string `json:"gas"`
	GasPrice          string `json:"gasPrice"`
	Hash              string `json:"hash"`
	To                string `json:"to"`
	Timestamp         uint   `json:"timeStamp,string"`
	TransactionIndex  int    `json:"transactionIndex,string"`
	Value             string `json:"value"`
	TxReceiptStatus   int    `json:"txreceipt_status,string"`
	IsError           int    `json:"isError,string"`
}

type TXResp struct {
	Message string `json:"message"`
	Result  []TX   `json:"result"`
	Status  string `json:"status"`
}

var (
	fConfigFile            string
	fDebug                 bool
	file                   string
	fInit, fWatch, fStatus bool
	fApiUrl                string
	fFile                  string
	fRefresh               bool
)

func init() {
	flag.BoolVar(&fDebug, "debug", false, "Debug mode")
	flag.StringVar(&fConfigFile, "config", "/etc/config.ini", "Configuration file")
	flag.StringVar(&file, "file", "", "File for export")
	flag.BoolVar(&fWatch, "watch", false, "Search for transactions for existing addresses")
	flag.BoolVar(&fInit, "init", false, "DB Init")
	flag.BoolVar(&fStatus, "status", false, "Show key statuses")
	flag.BoolVar(&fRefresh, "refresh", false, "Refresh data from database")
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(btcec.S256(), rand.Reader)
}

func Keccak256(in []byte) []byte {
	hash := sha3.NewKeccak256()

	hash.Write(in)
	return hash.Sum(nil)
}

func DbConnect(dbHost, dbName, dbUser, dbPass string) mysql.Conn {
	db := mysql.New("tcp", "", fmt.Sprintf("%s:3306", dbHost), dbUser, dbPass, dbName)
	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func CreateTable(db mysql.Conn) error {
	query := `CREATE TABLE ethkeys(
			    id INT NOT NULL AUTO_INCREMENT,
			    pub CHAR(40) NOT NULL,
			    used BOOL NOT NULL DEFAULT false,
			    completed BOOL NOT NULL DEFAULT false,
			    tx_metadata TEXT,
			    tx_value NUMERIC(32) DEFAULT 0,
			    received NUMERIC(32) DEFAULT 0,
			    started_ts TIMESTAMP DEFAULT '2018-01-01 00:00:00',
			    PRIMARY KEY(id));`

	_, _, err := db.Query(query)
	if err != nil {
		return err
	}

	query = `CREATE TABLE ethtx(
				id INT NOT NULL AUTO_INCREMENT,
				block_hash TEXT,
				block_number INT,
				confirmations INT,
				contract_address TEXT,
				cumulative_gas_used TEXT,
				from_addr TEXT,
				gas TEXT,
				gas_price TEXT,
				hash VARCHAR(66) UNIQUE NOT NULL,
				timestamp INT UNSIGNED,
				transaction_index INT,
				to_addr TEXT,
				value TEXT,
				tx_receipt_status INT,
				is_error INT,
				PRIMARY KEY(id));`

	_, _, err = db.Query(query)
	if err != nil {
		return err
	}

	return nil
}

func Store(db mysql.Conn, pub string) error {
	if fDebug {
		log.Printf("DB Store: %s\n", pub)
	}

	stmt, err := db.Prepare("INSERT INTO ethkeys(pub) VALUES(?)")
	if err != nil {
		return err
	}

	res, err := stmt.Run(pub)
	if err != nil {
		return err
	}

	if fDebug {
		log.Printf("DB Store: returned id:%d\n", res.InsertId())
	}

	return nil
}

func StoreTXs(db mysql.Conn, txs []TX) error {
	if fDebug {
		log.Printf("DB StoreTXs: %d records\n", len(txs))
	}

	stmt, err := db.Prepare(`INSERT IGNORE INTO ethtx(
		block_hash, block_number, confirmations, contract_address, cumulative_gas_used,
		from_addr,  gas, gas_price, hash, timestamp, transaction_index, to_addr, value,
		tx_receipt_status, is_error) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`)

	if err != nil {
		return err
	}

	for _, tx := range txs {
		if fDebug {
			log.Printf("DB StoreTXs: Storing tx hash:%s\n", tx.Hash)
		}
		res, err := stmt.Run(
			tx.BlockHash,
			tx.BlockNumber,
			tx.Confirmations,
			tx.ContractAddress,
			tx.CumulativeGasUsed,
			tx.From,
			tx.Gas,
			tx.GasPrice,
			tx.Hash,
			tx.Timestamp,
			tx.TransactionIndex,
			tx.To,
			tx.Value,
			tx.TxReceiptStatus,
			tx.IsError,
		)

		if err != nil {
			return err
		}

		if fDebug {
			log.Printf("DB StoreTXs: returned id:%d\n", res.InsertId())
		}

	}

	if fDebug {
		log.Printf("DB StoreTXs Done.\n")
	}

	return nil
}

func GetStoreStatus(db mysql.Conn) int {
	rows, _, err := db.Query("SELECT COUNT(*) as count FROM ethkeys WHERE used = false;")
	if err != nil {
		log.Fatal(err)
	}

	if len(rows) != 1 {
		return 0
	}

	return rows[0].Int(0)
}

func Prepend(in []byte, size int) []byte {
	if len(in) == size {
		return in
	}

	prefix := make([]byte, size-len(in))

	return append(prefix, in...)
}

func HttpQuery(url string) ([]byte, error) {
	if fDebug {
		log.Printf("HTTP: Doing HTTP Query: %s\n", url)
	}

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	if fDebug {
		log.Printf("HTTP: Got result %s\n", string(body))
	}

	return body, nil
}

func QueryEtherscan(addr string) (big.Int, error) {
	var balance BalanceResp
	var out big.Int

	url := fmt.Sprintf(
		"https://%s/api?module=account&action=balance&address=0x%s&tag=latest",
		fApiUrl,
		addr,
	)

	body, err := HttpQuery(url)
	if err != nil {
		return big.Int{}, err
	}

	err = json.Unmarshal(body, &balance)
	if err != nil {
		return big.Int{}, err
	}

	if fDebug {
		log.Printf("API: Addr: %s - Status: %s - Balance: %s\n", addr, balance.Message, balance.Result)
	}

	err = out.UnmarshalText([]byte(balance.Result))
	if err != nil {
		return big.Int{}, nil
	}

	if fDebug {
		log.Printf("API BigInt: %s\n", out.Text(10))
	}

	return out, nil
}

func QueryEtherscanTX(addr string) ([]TX, error) {
	var txs TXResp

	url := fmt.Sprintf(
		"https://%s/api?module=account&action=txlist&address=0x%s&startblock=0&endblock=99999999&sort=asc",
		fApiUrl,
		addr,
	)

	body, err := HttpQuery(url)
	if err != nil {
		return []TX{}, err
	}

	err = json.Unmarshal(body, &txs)
	if err != nil {
		return []TX{}, err
	}

	return txs.Result, nil
}

func UpdateValue(db mysql.Conn, pub string, value big.Int) error {
	stmt, err := db.Prepare("UPDATE ethkeys SET received = ? WHERE pub = ?")
	if err != nil {
		return err
	}

	_, err = stmt.Run(value.Text(10), pub)
	if err != nil {
		return err
	}

	if fDebug {
		log.Printf("UpdateValue pub:%s value:%s completed without error\n", pub, value.Text(10))
	}

	return nil
}

func Watch(db mysql.Conn) error {
	var current_received big.Int
	var pubkey string

	// https://api.etherscan.io/api?module=account&action=balance&address=&tag=latest&apikey=YourApiKeyToken
	if fDebug {
		log.Println("Watch()")
	}

	query := "SELECT pub, received FROM ethkeys WHERE tx_value > received AND NOW() < started_ts + INTERVAL 1 DAY AND completed = false AND received < tx_value;"

	if fRefresh {
		query = "SELECT pub, received FROM ethkeys WHERE used = true and completed = false;"
	}

	rows, _, err := db.Query(query)
	if err != nil {
		return err
	}

	if fDebug && len(rows) == 0 {
		log.Printf("No record to look after.")
	}

	for _, row := range rows {
		pubkey = row.Str(0)

		if fDebug {
			log.Printf("Looking for key 0x%s\n", pubkey)
		}

		value, err := QueryEtherscan(pubkey)
		if err != nil {
			return err
		}

		current_received.UnmarshalText([]byte(row.Str(1)))

		if fDebug {
			log.Printf("Current: pub:%s received:%s\n", pubkey, current_received.Text(10))

		}

		if value.Cmp(&current_received) != 0 || fRefresh {
			if fDebug {
				log.Printf("Storing new received value (%s) in database.\n", value.Text(10))
			}

			err := UpdateValue(db, pubkey, value)
			if err != nil {
				return err
			}

			if fDebug {
				log.Printf("Query for TX for 0x%s\n", pubkey)
			}

			txs, err := QueryEtherscanTX(pubkey)
			if err != nil {
				return err
			}

			err = StoreTXs(db, txs)
			if err != nil {
				return err
			}

		} else {
			if fDebug {
				log.Printf("No balance change.\n")
			}
		}
	}

	if fDebug {
		log.Println("Watch() done successfully!")
	}

	return nil
}

func ShowStatus(db mysql.Conn) error {
	rows, _, err := db.Query("SELECT id, pub, used, tx_value, received, started_ts FROM ethkeys WHERE completed = false ORDER BY ID ASC;")
	if err != nil {
		return err
	}

	for _, row := range rows {
		var started_ts_str string
		if row.Str(5) != "2018-01-01 00:00:00" {
			started_ts_str = fmt.Sprintf("started_ts:'%s'", row.Str(5))
		}

		log.Printf("id:%d 0x%s used:%v waited:%d received:%d %s\n", row.Int(0), row.Str(1), row.Bool(2), row.Int(3), row.Int(4), started_ts_str)
	}

	return nil
}

func main() {
	var db mysql.Conn
	var fd *os.File
	var err error

	flag.Parse()

	cfg, err := ini.Load(fConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v\n", err)
		os.Exit(1)
	}

	if !fDebug {
		fDebug = cfg.Section("general").Key("debug").MustBool(false)
	}

	pool_num := cfg.Section("keys").Key("num").MustInt(20)

	if fDebug {
		log.Printf("Key pool size is %d.\n", pool_num)
	}

	fApiUrl = cfg.Section("general").Key("api_url").MustString("api-ropsten.etherscan.io")

	if fDebug {
		log.Printf("Using API url: %s\n", fApiUrl)
	}

	if file == "" {
		file = cfg.Section("general").Key("file").MustString("./private-keys")
	}

	if fDebug {
		log.Printf("Using file: %s\n", file)
	}

	is_disabled := cfg.Section("db").Key("disabled").MustBool(false)
	if err != nil {
		fmt.Printf("Invalid value for disabled field: %s\n", cfg.Section("db").Key("disabled").String())
		os.Exit(1)
	}

	if is_disabled == false {
		db = DbConnect(
			cfg.Section("db").Key("host").String(),
			cfg.Section("db").Key("name").String(),
			cfg.Section("db").Key("user").String(),
			cfg.Section("db").Key("pass").String(),
		)
		defer db.Close()

		if fInit {
			err := CreateTable(db)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Println("Tables created.")
			}

			return
		}
	}

	if file != "" {
		fd, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Panic(err)
		}
		defer fd.Close()
	}

	if fStatus {
		ShowStatus(db)
		return
	}

	if fWatch {
		err := Watch(db)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	db_num := GetStoreStatus(db)

	if pool_num <= db_num {
		log.Printf("No need to insert new records (%d keys in DB).\n", db_num)
		return
	}

	required_num := pool_num - db_num

	if fDebug {
		log.Printf("Required to create %d new keys.\n", required_num)
	}

	for i := 0; i < required_num; i++ {
		key, err := GenerateKey()
		if err != nil {
			log.Panic(err)
		}

		var publickey bytes.Buffer
		// No 0x04 before hashing.
		// publickey.WriteByte(byte(0x04))
		publickey.Write(Prepend(key.PublicKey.X.Bytes(), 32))
		publickey.Write(Prepend(key.PublicKey.Y.Bytes(), 32))

		privatekey := Prepend(key.D.Bytes(), 32)

		hash := Keccak256(publickey.Bytes())

		if fDebug {
			log.Printf("Priv: %x\n", privatekey)
			log.Printf("Pub:  0x%x\n", hash[12:])
		}

		if fd != nil {
			fd.WriteString(fmt.Sprintf("0x%x;%x\n", hash[12:], privatekey))
		}

		if db != nil {
			Store(db, fmt.Sprintf("%x", hash[12:]))
		}
	}
}
