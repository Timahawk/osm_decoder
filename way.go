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

type MyLineString struct {
	Id     int64
	Tags   map[string]string
	Refs   []int64
	Coords geom.LineString
}

type MyPolygon struct {
	Id     int64
	Tags   map[string]string
	Refs   []int64
	Coords geom.Polygon
}

func decodeLinestring(way *pb.Way, strtable [][]byte) (string, MyLineString) {
	mls := MyLineString{way.GetId(), map[string]string{}, []int64{}, *geom.NewLineString(geom.XY)}

	keys := way.GetKeys()
	values := way.GetVals()
	for i, k := range keys {
		mls.Tags[string(strtable[k])] = string(strtable[values[i]])
	}

	refs := way.GetRefs()

	delta_ref := int64(0)
	liste_of_coords := []geom.Coord{}
	for _, k := range refs {
		delta_ref = k + delta_ref
		mls.Refs = append(mls.Refs, delta_ref)

		coord, ok := largeMapNode.Get(fmt.Sprint(delta_ref))
		if ok {
			liste_of_coords = append(liste_of_coords, coord.Coords.Coords())
		}
	}
	mls.Coords = *mls.Coords.MustSetCoords(liste_of_coords)
	return fmt.Sprint(way.GetId()), mls
}

func decodePolygon(way *pb.Way, strtable [][]byte) (string, MyPolygon) {
	mlp := MyPolygon{way.GetId(), map[string]string{}, []int64{}, *geom.NewPolygon(geom.XY)}

	keys := way.GetKeys()
	values := way.GetVals()
	for i, k := range keys {
		mlp.Tags[string(strtable[k])] = string(strtable[values[i]])
	}

	refs := way.GetRefs()

	delta_ref := int64(0)
	liste_of_coords := []geom.Coord{}
	for _, k := range refs {
		delta_ref = k + delta_ref
		mlp.Refs = append(mlp.Refs, delta_ref)

		coord, ok := largeMapNode.Get(fmt.Sprint(delta_ref))
		if ok {
			liste_of_coords = append(liste_of_coords, coord.Coords.Coords())
		}
	}
	mlp.Coords = *mlp.Coords.MustSetCoords([][]geom.Coord{liste_of_coords})
	return fmt.Sprint(way.GetId()), mlp
}

func DecodeWays(pg *pb.PrimitiveGroup, strtable [][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	ways := pg.GetWays()

	for _, way := range ways {

		refs := way.GetRefs()

		// This is duplicated within the decode functions.
		// TODO clean up
		decoded_refs := []int64{}
		delta_ref := int64(0)
		for _, ref := range refs {
			delta_ref = delta_ref + ref
			decoded_refs = append(decoded_refs, delta_ref)
		}

		if decoded_refs[0] != decoded_refs[len(refs)-1] {
			largeMapLineString.Set(decodeLinestring(way, strtable))
		} else {
			largeMapPolygon.Set(decodePolygon(way, strtable))
		}
	}
}

func (ls *MyLineString) SQLString() string {
	// Schreibe zum String wenn es einen Tag hat.
	if len(ls.Tags) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("(" + fmt.Sprint(ls.Id) + ", ('")

	j := 0
	len := len(ls.Tags)
	for key, value := range ls.Tags {

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
	str, err := encoder.Encode(&ls.Coords)
	if err != nil {
		log.Fatal(err)
	}

	sb.WriteString(str + "'))")
	return sb.String()
}

func (pg *MyPolygon) SQLString() string {
	// Schreibe zum String wenn es einen Tag hat.
	if len(pg.Tags) == 0 {
		return ""
	}

	var sb strings.Builder

	sb.WriteString("(" + fmt.Sprint(pg.Id) + ", ('")

	j := 0
	len := len(pg.Tags)
	for key, value := range pg.Tags {

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
	str, err := encoder.Encode(&pg.Coords)
	if err != nil {
		log.Fatal(err)
	}

	sb.WriteString(str + "'))")
	return sb.String()
}
