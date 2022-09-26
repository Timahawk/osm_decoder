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

type MyWayFeature struct {
	Feature    geojson.Feature
	LineString orb.LineString
	Polygon    orb.Polygon
	Refs       []int64
}

// type MyPolygonFeature struct {
// 	Feature    geojson.Feature
// 	LineString orb.LineString
// }

func DecodeWay(way *pb.Way, strtable [][]byte) (string, MyWayFeature) {

	feature := MyWayFeature{
		Feature: geojson.Feature{
			ID:         way.GetId(),
			Properties: map[string]interface{}{}}}

	keys := way.GetKeys()
	values := way.GetVals()
	for i, k := range keys {
		feature.Feature.Properties[string(strtable[k])] = strtable[values[i]]
	}

	refs := way.GetRefs()

	delta_ref := int64(0)

	ls := orb.LineString{}
	ring := orb.Ring{}

	for _, k := range refs {
		delta_ref = k + delta_ref

		feature.Refs = append(feature.Refs, delta_ref)

		point, ok := largeMapNode.Get(fmt.Sprint(delta_ref))
		if ok {
			ls = append(ls, point.Point)
			ring = append(ring, point.Point)
		}
	}

	if feature.Refs[0] == feature.Refs[len(feature.Refs)-1] {

		polygon := orb.Polygon{ring}

		feature.Feature.BBox = geojson.NewBBox(polygon.Bound())
		feature.Feature.Type = polygon.GeoJSONType()
		feature.Feature.Geometry = polygon

		return fmt.Sprint(feature.Feature.ID), feature
	} else {
		feature.Feature.BBox = geojson.NewBBox(ls.Bound())
		feature.Feature.Type = ls.GeoJSONType()
		feature.Feature.Geometry = ls

		return fmt.Sprint(feature.Feature.ID), feature
	}
}

/*
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
*/
func DecodeWays(pg *pb.PrimitiveGroup, strtable [][]byte, wg *sync.WaitGroup) {
	defer wg.Done()

	ways := pg.GetWays()

	for _, way := range ways {

		id, feature := DecodeWay(way, strtable)

		largeMapWays.Set(id, feature)
	}
}

func (f *MyWayFeature) SQLString() string {
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
	sb.WriteString("'), ST_GeomFromGeoJSON('{\"type\":\"" + f.Feature.Geometry.GeoJSONType() + "\", \"coordinates\":")

	str, err := json.Marshal(f.Feature.Geometry)
	if err != nil {
		log.Fatal(err)
	}

	sb.WriteString(string(str) + "}'))")
	return sb.String()
}

/*
func (f *MyLineString) SQLString() string {
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

/*
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
*/
