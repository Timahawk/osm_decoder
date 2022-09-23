package main

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/twpayne/go-geom/encoding/wkt"
)

func Setup() {
	file, err := os.Open("./testdata/andorra-latest.osm.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	defer file.Close()

	err = Parse(file)
	if err != nil {
		log.Fatal(err)
	}
}

func Test_decodeDenseNodes(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("625025")
	if !ok {
		log.Fatal("Not in dict!")
	}
	encoder := wkt.NewEncoder()
	str, err := encoder.Encode(&node.Coords)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, str, "POINT (1.5527243000000002 42.5142133)")
}

// Can only test on features with one Tag because the order is random.
// So test fail if result is correct because of Tag order.
func Test_SQLString(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("264932716")
	if !ok {
		log.Fatal("Not in dict!")
	}
	assert.Equal(t, node.SQLString(), "(264932716, ('\"created_by\"=>\"JOSM\"'), ST_GeomFromText('POINT (1.6483839 42.6004428)'))")
}
