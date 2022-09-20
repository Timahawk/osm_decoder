package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/jackc/pgx/v5"
)

type MyNode struct {
	id   int64
	lat  float64
	lon  float64
	tags map[string]string
}

func calc(num int64, pbs *PrimBlogSettings) float64 {
	return pbs.coordScale * float64(pbs.lonOffset+(pbs.granularity*num))
}

func decodeDenseNodes(pg *pb.PrimitiveGroup, st *pb.StringTable, pbs *PrimBlogSettings, conn *pgx.Conn) []*MyNode {

	// fmt.Println(pbs)

	MyNodes := []*MyNode{}

	strtable := st.GetS()

	densenodes := pg.GetDense()
	id := densenodes.GetId()
	lat := densenodes.GetLat()
	lon := densenodes.GetLon()
	tags := densenodes.GetKeysVals()

	delta_id := int64(0)
	delta_lat := int64(0)
	delta_lon := int64(0)

	// Counts where we at in the stringTable
	counter := 0

	var sb strings.Builder
	sb.WriteString("Insert INTO points (id, geom, tags) VALUES ")

	for i := 0; i < len(id); i++ {

		delta_id = id[i] + delta_id
		delta_lat = lat[i] + delta_lat
		delta_lon = lon[i] + delta_lon

		mn := MyNode{delta_id, calc(delta_lat, pbs), calc(delta_lon, pbs), map[string]string{}}

		for {
			if tags[counter] == 0 {
				counter += 1
				break
			} else {
				mn.tags[string(strtable[tags[counter]])] = string(strtable[tags[counter+1]])
				counter += 2
			}
		}

		MyNodes = append(MyNodes, &mn)

		// Schreibe zum String wenn es einen Tag hat.
		if len(mn.tags) != 0 {

			sb.WriteString(fmt.Sprintf("(%v, ST_GeomFromText('POINT(%v %v)'),'", mn.id, mn.lon, mn.lat))

			j := 0
			l := len(mn.tags)
			for key, value := range mn.tags {

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

	// Catch no Nodes to write.
	if sb.Len() > 30 {
		return MyNodes
	}
	str := sb.String()[:len(sb.String())-1]

	if ToDB {
		tx, _ := conn.Begin(context.Background())
		_, err := conn.Exec(context.Background(), str)
		if err != nil {
			log.Fatal("Exec:", err, str)
		}
		tx.Commit(context.Background())
	}
	return MyNodes
}
