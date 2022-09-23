package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"google.golang.org/protobuf/proto"
)

func Test_do(t *testing.T) {

	feature1 := pb.Tile_Feature{}
	feature2 := pb.Tile_Feature{}
	// feature3 := pb.Tile_Feature{}
	// feature4 := pb.Tile_Feature{}

	layer1 := pb.Tile_Layer{}

	str := "layer1"
	n := uint32(1)
	ext := uint32(4096)

	layer1.Version = &n
	layer1.Name = &str
	layer1.Features = append(layer1.GetFeatures(), &feature1, &feature2)
	layer1.Keys = []string{"la", "le", "lu"}
	// layer1.Values = []string{"la", "le", "lu"}
	layer1.Extent = &ext
	layer2 := pb.Tile_Layer{}
	layer3 := pb.Tile_Layer{}

	mytile := pb.Tile{}
	mytile.Layers = append(mytile.GetLayers(), &layer1, &layer2, &layer3)

	fmt.Println(mytile.String())
	t.Fail()
}

func Test_Decodeing(t *testing.T) {
	file, err := os.ReadFile("./data/11340.vector.pbf")
	if err != nil {
		log.Fatalln("Error reading file:", err)
	}
	//defer file.Close()

	tile := pb.Tile{}
	err = proto.Unmarshal(file, &tile)
	if err != nil {
		log.Fatalln(err)
	}
	for _, layer := range tile.GetLayers() {
		fmt.Println(layer.GetName())
	}
	t.Fail()
}

// func Test_DecodeBrotli(t *testing.T) {
// 	file, err := os.Open("./data/aux_11_1088_709.pbf")
// 	if err != nil {
// 		log.Fatalln("Error Openening:", err)
// 	}
// 	defer file.Close()

// 	r := brotli.NewReader(file)

// 	res := make([]byte, 1)

// 	n, err := r.Read(res)
// 	if err != nil {
// 		log.Fatalln("Error Reading:", err)
// 	}
// 	fmt.Println(n)

// 	err = os.WriteFile("./data/aux_decompressed.pbf", res, 0666)
// 	if err != nil {
// 		log.Fatalln("Error Reading:", err)
// 	}
// 	t.Fail()
// }
