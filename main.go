package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	"fmt"
	"log"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

var (
	debug          bool
	num            int
	dbHost         string
	dbUser, dbPass string
	dbName         string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.StringVar(&dbHost, "db-host", "localhost", "DB Hostname")
	flag.StringVar(&dbName, "db-name", "", "DB Name (mandatory)")
	flag.StringVar(&dbUser, "db-user", "", "DB User")
	flag.StringVar(&dbPass, "db-pass", "", "DB Pass")
	flag.IntVar(&num, "num", 1, "Num of keys to generate")
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(btcec.S256(), rand.Reader)
}

func Keccak256(in []byte) []byte {
	hash := sha3.NewKeccak256()

	hash.Write(in)
	return hash.Sum(nil)
}

func DbConnect() *sql.DB {
	connectString := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPass, dbHost, dbName)

	if debug {
		log.Printf("Storing using %s", connectString)
	}

	db, err := sql.Open("mysql", connectString)
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func Store(db *sql.DB, pub, priv string) bool {
	// Table used:
	// CREATE TABLE ethkeys(
	//   id MEDIUMINT NOT NULL AUTO_INCREMENT,
	//   pub CHAR(40) NOT NULL,
	//   priv CHAR(64) NOT NULL,
	//   PRIMARY KEY(id));

	stmt, err := db.Prepare("INSERT INTO ethkeys(pub, priv) VALUES(?, ?)")
	if err != nil {
		log.Fatal(err)
	}

	res, err := stmt.Exec(pub, priv)
	if err != nil {
		log.Fatal(err)
	}

	lastId, err := res.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Printf("Record: %d\n", lastId)
	}

	return true
}

func Prepend(in []byte, size int) []byte {
	if len(in) == size {
		return in
	}

	prefix := make([]byte, size-len(in))

	return append(prefix, in...)
}

func main() {
	flag.Parse()

	var db *sql.DB

	if dbName != "" {
		db = DbConnect()
		defer db.Close()
	}

	for i := 0; i < num; i++ {
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

		if debug {
			log.Printf("Priv: %x\n", privatekey)
			log.Printf("Pub:  0x%x\n", hash[12:])
		}

		if dbName != "" {
			Store(db, fmt.Sprintf("%x", hash[12:]), fmt.Sprintf("%x", privatekey))
		}
	}
}
