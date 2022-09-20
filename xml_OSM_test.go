package main

import (
	"fmt"
	"log"
	"testing"
)

func TestOsmFile(t *testing.T) {
	osm := NewOsmFile()

	err := osm.Read("data/building_Streets.osm")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Ways:", len(osm.Ways), "Nodes:", len(osm.Nodes))
}
