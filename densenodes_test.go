package main

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Setup() {
	file, err := os.Open("./testdata/andorra-latest.osm.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	defer file.Close()

	err = Parse(file)
	if err != nil {
		log.Fatal(err)
	}
}

func Test_NodetoGEOJSON(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("625025")
	if !ok {
		log.Fatal("Not in dict!")
	}
	var str []byte
	str, err := json.Marshal(node.Feature)
	if err != nil {
		log.Fatal("Not in dict!")
	}
	assert.Equal(t,
		"{\"id\":625025,\"type\":\"Feature\",\"bbox\":[1.5527243000000002,42.5142133],\"geometry\":{\"type\":\"Point\",\"coordinates\":[1.5527243000000002,42.5142133]},\"properties\":null}",
		string(str))
}

/*
// Can only test on features with one Tag because the order is random.
// So test fail if result is correct because of Tag order.
func Test_SQLString(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("264932716")
	if !ok {
		log.Fatal("Not in dict!")
	}
	assert.Equal(t, node.SQLString(), "(264932716, ('\"created_by\"=>\"JOSM\"'), ST_GeomFromText('POINT (1.6483839 42.6004428)'))")
}

func Test_GeoJSON_go_geom(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("4698145789")
	if !ok {
		log.Fatal("Not in dict!")
	}
	node2, ok := largeMapNode.Get("625025")
	if !ok {
		log.Fatal("Not in dict!")
	}

	newmap := map[string]interface{}{}
	for key, value := range node.Tags {
		newmap[key] = value
	}

	mn := geojson.Feature{
		ID:         fmt.Sprint(node.Id),
		BBox:       node.Coords.Bounds(),
		Geometry:   &node.Coords,
		Properties: newmap}

	mn2 := geojson.Feature{
		ID:         fmt.Sprint(node2.Id),
		BBox:       node2.Coords.Bounds(),
		Geometry:   &node2.Coords,
		Properties: map[string]interface{}{}}

	// bytes, err := mn2.MarshalJSON()
	// if err != nil {
	// 	log.Fatal("Not in dict!")
	// }
	bounds := geom.NewBounds(geom.XY)
	bounds.SetCoords(mn.Geometry.FlatCoords(), mn2.Geometry.FlatCoords())
	features := []*geojson.Feature{&mn, &mn2}
	fc := geojson.FeatureCollection{
		BBox:     bounds,
		Features: features}
	gj, err := fc.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, "", string(gj))

}

func Test_GeoJSON_orb(t *testing.T) {
	Setup()

	node, ok := largeMapNode.Get("4698145789")
	if !ok {
		log.Fatal("Not in dict!")
	}

	newmap := map[string]interface{}{}
	for key, value := range node.Tags {
		newmap[key] = value
	}

	coords := node.Coords.FlatCoords()
	point := orb.Point{coords[0], coords[1]}
	mn := orbgj.Feature{
		ID:   fmt.Sprint(node.Id),
		Type: "Point",
		// BBox:       node.Coords.Bounds(),
		Geometry:   point,
		Properties: newmap}

	fc := orbgj.FeatureCollection{
		Features: []*orbgj.Feature{&mn},
	}
	data, err := fc.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	println(string(data))
	assert.Equal(t, 1, 2)
}

func BenchmarkSingleLoop(b *testing.B) {
	os.Stdout, _ = os.Open(os.DevNull)
	Setup()
	cnt := 0

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		valuesNodes := largeMapNode.IterBuffered()
		for tuple := range valuesNodes {
			if val, ok := tuple.Val.Tags["natural"]; ok {
				if val == "tree" {
					cnt++
				}
			}
		}
	}
	log.Println("Number of trees: ", cnt)
}

func BenchmarkParallelLoop(b *testing.B) {

	os.Stdout, _ = os.Open(os.DevNull)
	Setup()
	cnt := 0

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		largeMapNode.IterCb(func(key string, v MyNode) {
			if val, ok := v.Tags["natural"]; ok {
				if val == "tree" {
					cnt++
				}
			}
		})
	}
	log.Println("Number of trees: ", cnt)
}
*/
