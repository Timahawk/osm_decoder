package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// var db *badger.DB
// var Conn *pgx.Conn

var all = true

var ToDB_Points = all
var ToDB_LineString = all
var ToDB_Polygons = all

var FailCnt = 0

var connstring = "postgresql://postgres:postgres@localhost:5432/osmconverter"
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

var Pool *pgxpool.Pool

func main() {

	Pool, _ = pgxpool.New(context.Background(), connstring)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	defer Pool.Close()

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

	// file, err := os.Open("D:/Downloads/rheinland-pfalz-latest.osm.pbf")
	file, err := os.Open("D:/Downloads/andorra-latest.osm.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	defer file.Close()

	// db = SetupDB()
	// defer db.Close()

	LoopOverFile(file)

	fmt.Println("Processed Finished in:", time.Since(start))
}
