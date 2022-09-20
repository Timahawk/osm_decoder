package main

import (
	"log"

	"github.com/dgraph-io/badger/v3"
)

func SetupDB() *badger.DB {
	// Open the Badger database located in the /tmp/badger directory.
	// It will be created if it doesn't exist.

	opt := badger.DefaultOptions("").WithInMemory(true)

	db, err := badger.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
