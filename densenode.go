package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	pb "github.com/Timahawk/osm_file_decoder/proto"
)

type MyNode struct {
	Id int64
	// Lat    float64
	// Lon    float64
	Tags   map[string]string
	Coords Coord
}

// calc converts the value stored in Lat/Lon into valid WGS84 coordiantes
// If for some reason lonOffset & latOffset are different, this will make
// all latitude results incorrect
func calc(num int64, pbs *primBlockSettings) float64 {
	return pbs.coordScale * float64(pbs.lonOffset+(pbs.granularity*num))
}

// decodeDenseNodes loops over all Nodes within a Primitive Group and decodes them.
// If flag is set, it also writes them to the DB.
func decodeDenseNodes(pg *pb.PrimitiveGroup, st *pb.StringTable, pbs *primBlockSettings, wg *sync.WaitGroup) {
	defer wg.Done()
	// MyNodes := []*MyNode{}

	strtable := st.GetS()
	// Counts where we at in the stringTable
	counter := 0

	densenodes := pg.GetDense()
	Id := densenodes.GetId()
	Lat := densenodes.GetLat()
	Lon := densenodes.GetLon()
	tags := densenodes.GetKeysVals()

	delta_id := int64(0)
	delta_Lat := int64(0)
	delta_Lon := int64(0)

	var sb strings.Builder
	sb.WriteString("Insert INTO points (id, geom, tags) VALUES ")

	for i := 0; i < len(Id); i++ {

		delta_id = Id[i] + delta_id
		delta_Lat = Lat[i] + delta_Lat
		delta_Lon = Lon[i] + delta_Lon

		mn := MyNode{delta_id, map[string]string{}, Coord{calc(delta_Lat, pbs), calc(delta_Lon, pbs)}}

		for {
			if tags[counter] == 0 {
				counter += 1
				break
			} else {
				mn.Tags[string(strtable[tags[counter]])] = string(strtable[tags[counter+1]])
				counter += 2
			}
		}

		// MyNodes = append(MyNodes, &mn)
		largeMapNode.Set(fmt.Sprint(delta_id), mn)

		// Schreibe zum String wenn es einen Tag hat.
		if len(mn.Tags) != 0 {

			sb.WriteString(fmt.Sprintf("(%v, ST_GeomFromText('POINT(%v %v)'),'", mn.Id, mn.Coords.Lon, mn.Coords.Lat))

			j := 0
			l := len(mn.Tags)
			for key, value := range mn.Tags {

				// Das weil das sonst Escaped.
				// TODO nicht weglasssen sondern Umformen
				if strings.ContainsRune(value, '\'') {
					j += 1
					continue
				}

				// Muss wegen dem Komma am Ende.
				if j+1 == l {
					sb.WriteString(fmt.Sprintf("%q=>%q", key, value))
				} else {
					sb.WriteString(fmt.Sprintf("%q=>%q, ", key, value))
				}
				j += 1
			}
			sb.WriteString("'),")
		}

	}

	if ToDB_Points {
		// Catch no Nodes to write.
		if sb.Len() > 50 {

			str := sb.String()[:len(sb.String())-1]
			err := Insert(str)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
