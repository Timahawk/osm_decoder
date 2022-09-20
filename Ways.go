package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/jackc/pgx/v5"
)

type Coord struct {
	lat float64
	lon float64
}
type MyWay struct {
	Id     int64
	Tags   map[string]string
	Refs   []int64
	Coords []Coord
	Type   string
}

func (mw *MyWay) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("(%v, '", mw.Id))

	j := 0
	l := len(mw.Tags)
	for key, value := range mw.Tags {

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
	sb.WriteString("', array[")

	j = 0
	l = len(mw.Refs)
	for _, ref := range mw.Refs {

		// Muss wegen dem Komma am Ende.
		if j+1 == l {
			sb.WriteString(fmt.Sprintf("%v", ref))
		} else {
			sb.WriteString(fmt.Sprintf("%v,", ref))
		}
		j += 1
	}

	if mw.Type == "LineString" {
		sb.WriteString("], ST_GeomFromText('LINESTRING(")
	} else {
		sb.WriteString("], ST_GeomFromText('POLYGON((")
	}

	// Coords not zugeordnet
	counter := 0
	num := len(mw.Coords)
	for i, coord := range mw.Coords {

		if coord.lat == 0 {
			continue
		}
		// Muss wegen dem Komma am Ende.
		if i+1 == num {
			sb.WriteString(fmt.Sprintf("%v %v", coord.lon, coord.lat))
			counter += 1
		} else {
			sb.WriteString(fmt.Sprintf("%v %v,", coord.lon, coord.lat))
			counter += 1
		}
	}

	// TODO das hier muss alles gefixt werden
	// TODO aber möglichst bald...

	if counter == 0 {
		// fmt.Println("Not zugeordnet at all", mw.Id, mw.Refs, mw.Coords)
		return ""
	} else if counter == 1 {
		// fmt.Println("Nur eine Coord zugeordnet", mw.Id, mw.Refs, mw.Coords)
		return ""

		// Falls geom am Anfang aber dann nur noch nullen:
	} else if sb.String()[sb.Len()-1] == ',' && counter > 1 {
		// fmt.Println("Komma am Ende", mw.Id, mw.Refs, mw.Coords)
		str := sb.String()[:sb.Len()-1]
		sb.Reset()
		sb.WriteString(str)
		sb.WriteString(" ")
	}

	if mw.Type == "LineString" {
		sb.WriteString(")', 4326) ), ")
	} else {
		sb.WriteString("))', 4326) ), ")
	}

	return sb.String()
}

func DecodeWays(pg *pb.PrimitiveGroup, st *pb.StringTable, largeMap map[int64]Coord, conn *pgx.Conn) []*MyWay {

	myways := []*MyWay{}

	strtable := st.GetS()
	delta_id := int64(0)
	ways := pg.GetWays()

	var sb strings.Builder

	sb.WriteString("INSERT INTO Lines VALUES ")
	// fmt.Println(sb.Len())

	for _, way := range ways {

		delta_id += way.GetId()
		mw := &MyWay{delta_id, map[string]string{}, []int64{}, []Coord{}, ""}

		keys := way.GetKeys()
		values := way.GetVals()
		refs := way.GetRefs()

		for i, k := range keys {
			mw.Tags[string(strtable[k])] = string(strtable[values[i]])
		}

		// for key, value := range largeMap {
		// 	fmt.Println("LARGE MAP", key, value)
		// }

		delta_id := int64(0)
		for _, k := range refs {
			delta_id = k + delta_id

			mw.Refs = append(mw.Refs, delta_id)
			mw.Coords = append(mw.Coords, largeMap[delta_id])
		}

		if mw.Refs[0] == mw.Refs[len(mw.Refs)-1] {
			mw.Type = "Polygon"
		} else {
			mw.Type = "LineString"
		}

		myways = append(myways, mw)

		if mw.Type == "LineString" {

			str := mw.String()
			if str == "" {
				failCnt += 1
			}
			sb.WriteString(str)
		}
	}
	var str string
	if sb.Len() > 30 {
		str = sb.String()[:len(sb.String())-2]
	} else {
		// fmt.Println(sb.String())
		return myways
	}

	if ToDB_LineString {

		tx, err := conn.Begin(context.Background())
		if err != nil {
			log.Fatal("Begin:", err, str)
		}
		_, err = conn.Exec(context.Background(), str)
		if err != nil {
			log.Fatal("Exec:", err, str[:50000])
		}
		err = tx.Commit(context.Background())
		if err != nil {
			log.Fatal("Commit:", err, str)
		}
	}

	return myways
}
