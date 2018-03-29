package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"

	// "database/sql"
	// _ "github.com/go-sql-driver/mysql"

	"github.com/ziutek/mymysql/mysql"
	_ "github.com/ziutek/mymysql/native" // Native engine

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)

var (
	debug          bool
	num            int
	dbHost         string
	dbUser, dbPass string
	dbName         string
	file           string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "Debug mode")
	flag.StringVar(&dbHost, "db-host", "localhost", "DB Hostname")
	flag.StringVar(&dbName, "db-name", "", "DB Name (mandatory)")
	flag.StringVar(&dbUser, "db-user", "", "DB User")
	flag.StringVar(&dbPass, "db-pass", "", "DB Pass")
	flag.IntVar(&num, "num", 1, "Num of keys to generate")
	flag.StringVar(&file, "file", "", "File for export")
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(btcec.S256(), rand.Reader)
}

func Keccak256(in []byte) []byte {
	hash := sha3.NewKeccak256()

	hash.Write(in)
	return hash.Sum(nil)
}

func DbConnect() mysql.Conn {
	db := mysql.New("tcp", "", fmt.Sprintf("%s:3306", dbHost), dbUser, dbPass, dbName)
	err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}

	return db
}

func Store(db mysql.Conn, pub, priv string) bool {
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

	res, err := stmt.Run(pub, priv)
	if err != nil {
		log.Fatal(err)
	}

	if debug {
		log.Printf("Record: %d\n", res.InsertId())
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

	var db mysql.Conn
	var fd *os.File
	var err error

	if dbName != "" {
		db = DbConnect()
		defer db.Close()
	}

	if file != "" {
		fd, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0600)
		if err != nil {
			log.Panic(err)
		}
		defer fd.Close()
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

		if fd != nil {
			fd.WriteString(fmt.Sprintf("0x%x;%x\n", hash[12:], privatekey))
		}

		if dbName != "" {
			Store(db, fmt.Sprintf("%x", hash[12:]), fmt.Sprintf("%x", privatekey))
		}
	}
}
