package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/twpayne/go-geom"
	"github.com/twpayne/go-geom/encoding/wkt"
)

type MyNode struct {
	Id     int64
	Tags   map[string]string
	Coords geom.Point
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

		mn := MyNode{delta_id, map[string]string{}, *geom.NewPoint(geom.XY).MustSetCoords([]float64{calc(delta_Lon, pbs), calc(delta_Lat, pbs)}).SetSRID(4326)}

		for {
			if tags[counter] == 0 {
				counter += 1
				break
			} else {
				mn.Tags[string(strtable[tags[counter]])] = string(strtable[tags[counter+1]])
				counter += 2
			}
		}

		largeMapNode.Set(fmt.Sprint(delta_id), mn)
	}
}

func (n *MyNode) SQLString() string {
	// Schreibe zum String wenn es einen Tag hat.
	if len(n.Tags) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("(" + fmt.Sprint(n.Id) + ", ('")

	j := 0
	len := len(n.Tags)
	for key, value := range n.Tags {

		// Das weil das sonst Escaped.
		// TODO nicht weglasssen sondern Umformen
		if strings.ContainsRune(value, '\'') {
			j += 1
			continue
		}

		// Muss wegen dem Komma am Ende.
		if j+1 == len {
			sb.WriteString(fmt.Sprintf("%q=>%q", key, value))
		} else {
			sb.WriteString(fmt.Sprintf("%q=>%q, ", key, value))
		}
		j += 1
	}
	sb.WriteString("'), ST_GeomFromText('")

	encoder := wkt.NewEncoder()
	str, err := encoder.Encode(&n.Coords)
	if err != nil {
		log.Fatal(err)
	}

	sb.WriteString(str + "'))")
	return sb.String()
}
