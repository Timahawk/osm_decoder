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

var connstring = "postgresql://postgres:postgres@localhost:5432/osmconverter"

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var upload = flag.Bool("upload", false, "Upload data to postgres")

func main() {

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
	file, err := os.Open("./testdata/andorra-latest.osm.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	defer file.Close()

	err = Parse(file)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Parsing done in:", time.Since(start))

	if *upload {
		start := time.Now()
		pool, err := pgxpool.New(context.Background(), connstring)
		if err != nil {
			log.Fatal(err)
		}
		defer pool.Close()

		toDB(pool)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Uploading done in :", time.Since(start))
	}

	fmt.Println("Everything finished in:", time.Since(start))
}
