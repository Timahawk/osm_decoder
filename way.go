package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	pb "github.com/Timahawk/osm_file_decoder/proto"
)

type Coord struct {
	Lat float64
	Lon float64
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

	counter := 0 // Number of Coords in Result String.
	num := len(mw.Coords)
	for i, coord := range mw.Coords {

		if coord.Lat == 0 {
			continue
		}
		// Muss wegen dem Komma am Ende.
		if i+1 == num {
			sb.WriteString(fmt.Sprintf("%v %v", coord.Lon, coord.Lat))
			counter += 1
		} else {
			sb.WriteString(fmt.Sprintf("%v %v,", coord.Lon, coord.Lat))
			counter += 1
		}
	}

	// TODO das hier muss alles gefixt werden
	// TODO aber möglichst bald...

	if counter == 0 {
		fmt.Println("Not zugeordnet at all", mw.Id, mw.Refs, mw.Coords)
		return ""
	} else if counter == 1 {
		fmt.Println("Nur eine Coord zugeordnet", mw.Id, mw.Refs, mw.Coords)
		return ""
	} else if mw.Type == "Polygon" && counter <= 3 {
		fmt.Println("Zu wenig Coords zugeordnet", mw.Id, mw.Refs, mw.Coords)
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

func DecodeWays(pg *pb.PrimitiveGroup, st *pb.StringTable, wg *sync.WaitGroup) {
	defer wg.Done()
	// myways := []*MyWay{}

	strtable := st.GetS()
	ways := pg.GetWays()

	var sb strings.Builder
	var sbp strings.Builder

	sb.WriteString("INSERT INTO Lines VALUES ")
	sbp.WriteString("INSERT INTO Polygons VALUES ")
	// fmt.Println(sb.Len())

	for _, way := range ways {

		mw := &MyWay{way.GetId(), map[string]string{}, []int64{}, []Coord{}, ""}

		keys := way.GetKeys()
		values := way.GetVals()
		refs := way.GetRefs()

		for i, k := range keys {
			mw.Tags[string(strtable[k])] = string(strtable[values[i]])
		}

		delta_ref := int64(0)
		for _, k := range refs {
			delta_ref = k + delta_ref

			mw.Refs = append(mw.Refs, delta_ref)
			coord, _ := largeMapNode.Get(fmt.Sprint(delta_ref))
			mw.Coords = append(mw.Coords, coord.Coords)
		}

		if mw.Refs[0] == mw.Refs[len(mw.Refs)-1] {
			mw.Type = "Polygon"
		} else {
			mw.Type = "LineString"
		}

		// myways = append(myways, mw)

		var str string
		if len(mw.Tags) > 0 {
			str = mw.String()
		} else {
			str = ""
		}
		if str == "" {
			FailCnt += 1
			// fmt.Println("Failcnt", failCnt)
		}
		if mw.Type == "LineString" {
			sb.WriteString(str)
		}
		if mw.Type == "Polygon" {
			sbp.WriteString(str)
		}
	}

	if ToDB_LineString {
		// Catch no Nodes to write.
		if sb.Len() > 30 {
			str := sb.String()[:len(sb.String())-2]
			err := Insert(str)
			if err != nil {
				log.Fatal(err)
			}
		}

	}

	if ToDB_Polygons {
		// Catch no Nodes to write.
		if sbp.Len() > 30 {
			str := sbp.String()[:len(sbp.String())-2]
			err := Insert(str)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	// return myways
}
