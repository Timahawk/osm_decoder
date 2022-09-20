package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strings"
)

type Coords struct {
	Lat  float64
	Long float64
}

type NodeX struct {
	Id   int64   `xml:"id,attr"`
	Lat  float64 `xml:"lat,attr"`
	Long float64 `xml:"lon,attr"`
}

type Ref struct {
	Id int64 `xml:"ref,attr"`
}

type Tag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

type WayX struct {
	Id     int64 `xml:"id,attr"`
	Refs   []Ref `xml:"nd"`
	Tags   []Tag `xml:"tag"`
	Coords []Coords
}

func (w *WayX) FormatCoords() string {
	var sb strings.Builder

	if w.Coords[0] == w.Coords[len(w.Coords)-1] {
		sb.WriteString("POLYGON((")
	} else {
		sb.WriteString("LINESTRING((")
	}

	for _, Coords := range w.Coords[:len(w.Coords)-1] {
		sb.WriteString(fmt.Sprintf("%f %f,", Coords.Long, Coords.Lat))
	}
	sb.WriteString(fmt.Sprintf("%f %f", w.Coords[len(w.Coords)-1].Long, w.Coords[len(w.Coords)-1].Lat))

	sb.WriteString("))")

	return sb.String()
}

type OsmFile struct {
	Nodes   []NodeX `xml:"node"`
	Ways    []WayX  `xml:"way"`
	NodesKV map[int64]Coords
}

func NewOsmFile() *OsmFile {
	var osm OsmFile
	osm.NodesKV = make(map[int64]Coords)
	return &osm
}

func (osm *OsmFile) Read(path string) error {
	log.Println("Start Reading XML File")

	// file, err := os.ReadFile("building_Streets.osm")
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("os.ReadFile Error, %d", err)
	}

	err = xml.Unmarshal(file, osm)
	if err != nil {
		return fmt.Errorf("xml.Unmarshal Error, %d", err)
	}

	for _, Node := range osm.Nodes {
		osm.NodesKV[Node.Id] = Coords{Node.Lat, Node.Long}
	}

	for idx, way := range osm.Ways {
		for _, ref := range way.Refs {
			// range is copy by value
			// way.Coords = append(way.Coords, nodesKV[ref.Id])
			osm.Ways[idx].Coords = append(osm.Ways[idx].Coords, osm.NodesKV[ref.Id])
		}
	}
	log.Println("Finished reading XML")
	return nil
}
