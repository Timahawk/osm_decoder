package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
)

type MyPointFeature struct {
	Feature geojson.Feature
	// this is because you cannot acsess the coords of a Feature
	Point orb.Point
}

// calc converts the value stored in Lat/Lon into valid WGS84 coordiantes
// If for some reason lonOffset & latOffset are different, this will make
// all latitude results incorrect
func calc(num int64, pbs *primBlockSettings) float64 {
	return pbs.coordScale * float64(pbs.lonOffset+(pbs.granularity*num))
}

// decodeDenseNodes loops over all Nodes within a Primitive Group and decodes them.
// If flag is set, it also writes them to the DB.
func decodeDenseNodes(pg *pb.PrimitiveGroup, strtable [][]byte, pbs *primBlockSettings, wg *sync.WaitGroup) {
	defer wg.Done()

	counter := 0 // Counts where we at in the stringTable

	densenodes := pg.GetDense()
	Id := densenodes.GetId()
	Lat := densenodes.GetLat()
	Lon := densenodes.GetLon()
	tags := densenodes.GetKeysVals()

	delta_id := int64(0)
	delta_Lat := int64(0)
	delta_Lon := int64(0)

	for i := 0; i < len(Id); i++ {

		delta_id = Id[i] + delta_id
		delta_Lat = Lat[i] + delta_Lat
		delta_Lon = Lon[i] + delta_Lon

		p := orb.Point{calc(delta_Lon, pbs), calc(delta_Lat, pbs)}

		feature := MyPointFeature{
			Feature: geojson.Feature{
				ID:         delta_id,
				BBox:       []float64{p.X(), p.Y()},
				Type:       p.GeoJSONType(),
				Geometry:   p,
				Properties: geojson.Properties{}},
			Point: p}

		for {
			if tags[counter] == 0 {
				counter += 1
				break
			} else {
				feature.Feature.Properties[string(strtable[tags[counter]])] = strtable[tags[counter+1]]
				counter += 2
			}
		}

		largeMapNode.Set(fmt.Sprint(delta_id), feature)
	}
}

func (f *MyPointFeature) SQLString() string {
	// Schreibe zum String wenn es einen Tag hat.
	if len(f.Feature.Properties) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("(" + fmt.Sprint(f.Feature.ID) + ", ('")

	j := 0
	len := len(f.Feature.Properties)
	for key, value := range f.Feature.Properties {

		value2 := fmt.Sprintf("%s", value)
		value2 = strings.Replace(value2, "'", "''", -1)

		// Muss wegen dem Komma am Ende.
		if j+1 == len {
			sb.WriteString(fmt.Sprintf("%q=>%q", key, value2))
		} else {
			sb.WriteString(fmt.Sprintf("%q=>%q, ", key, value2))
		}
		j += 1
	}
	sb.WriteString("'), ST_GeomFromGeoJSON('{\"type\":\"Point\", \"coordinates\":")

	str, err := json.Marshal(f.Point)
	if err != nil {
		log.Fatal(err)
	}

	sb.WriteString(string(str) + "}'))")
	return sb.String()
}
