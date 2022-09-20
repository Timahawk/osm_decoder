package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/jackc/pgx/v5"
)

// var db *badger.DB
// var Conn *pgx.Conn

var connstring = "postgresql://postgres:postgres@localhost:5432/osmconverter"
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {

	conn, err := pgx.Connect(context.Background(), connstring)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(context.Background())

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	start := time.Now()

	file, err := os.Open("D:/Master/Masterarbeit/data/karlsruhe-regbez-latest.osm.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	defer file.Close()

	// db = SetupDB()
	// defer db.Close()

	LoopOverFile(file, conn)

	fmt.Println("Processed Finished in:", time.Since(start))
}
