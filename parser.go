package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"

	"google.golang.org/protobuf/proto"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	cmap "github.com/orcaman/concurrent-map/v2"
)

// TODO switch to "github.com/orcaman/concurrent-map" for concurrency.
// var largeMapNode = make(map[int64]MyNode)
// var largeMapWays = make(map[int64]MyWay)

var largeMapNode = cmap.New[MyNode]()
var largeMapWays = cmap.New[MyWay]()

// var largeMapRelations = make(map[int64]MyRelation)

type primBlockSettings struct {
	granularity int64
	latOffset   int64
	lonOffset   int64
	coordScale  float64
}

// LoopOverFile loops over the complete .pbf file.
// The file is build up of BlobHeaders and Blobs
// Blobs can be HeaderBlock or an Primitive Block.
// BlobHeader und Blob kommen aus dem "fileformat.proto"
// HeaderBlock und PrimitiveBlock kommen aus "osmformat.proto"
func LoopOverFile(file *os.File) error {

	// Das ist ein BlobHeader
	blobHeader, err := extractBlobHeader(file)
	if err != nil {
		log.Fatalln("Error reading the first BlobHeader")
	}

	fmt.Println("First Header Blob:")
	fmt.Printf("\ttype: %v\n\tindexdata: %v\n\tdatasize: %v\n",
		blobHeader.GetType(),
		blobHeader.GetIndexdata(),
		blobHeader.GetDatasize())

	// Das  ist der Blob der den HeaderBlock enthält
	blob, err := extractBlob(blobHeader, file)
	if err != nil {
		log.Fatalln("Error reading the HeaderBlock")
	}
	// Ist einmalig im Datensatz. Deshalb außerhalb des loops.
	headerBlock := pb.HeaderBlock{}
	err = proto.Unmarshal(blob, &headerBlock)
	if err != nil {
		log.Fatalf("UnmarshalBlob error, %v", err)
	}
	fmt.Println("The Header Block:")
	fmt.Printf("\tBbox: %v\n\tRequiredFeatures: %v\n\tOptinalFeatures: %v\n\tWritingProgramm: %v\n",
		headerBlock.GetBbox(),
		headerBlock.GetRequiredFeatures(),
		headerBlock.GetOptionalFeatures(),
		headerBlock.GetWritingprogram())

	// testfile, _ := os.Create("testfile2.csv")
	// defer testfile.Close()
	// w := csv.NewWriter(testfile)

	// wayfile, _ := os.Create("wayfile.csv")
	// defer wayfile.Close()
	// waywriter := csv.NewWriter(wayfile)

	i := 0              // Numbers of Primitive Groups.
	counter_simple := 0 // DenseNodes without tag
	counter_tagged := 0 // DenseNodes with tag
	counter_LineString := 0
	counter_Polygon := 0

	c_dense, c_node, c_way, c_relation := 0, 0, 0, 0
	var wg sync.WaitGroup
	// Hier gehts quasi richtig los mit den Daten
	// Im ersten Block sind aber nur DenseNodes
	// DenseNodes sind nicht ein Struct mit jeweils Lat Long ID undsowas,
	// Sondern eine Liste die dann zusammengebastelt werden muss.
	// Die Eigenschaften kommen dann aus dem String Tabel
	// Ist alles ein wenig umständlich gemacht...
	// Ab hier wird dann solange über abwechselnd BlobHeader Blob geloopt bis error.
	// In den Blobs sind nur noch Blocks!
	for {

		blobHeader, err := extractBlobHeader(file)
		if err != nil {
			log.Println("Reached end of file after reading", i, "Blocks.")
			break
		}

		Block, err := extractBlob(blobHeader, file)
		if err != nil {
			fmt.Printf("Error reading Blob2\n")
			break
		}

		primitiveBlock := pb.PrimitiveBlock{}
		err = proto.Unmarshal(Block, &primitiveBlock)
		if err != nil {
			fmt.Printf("UnmarshalBlob error, %v\n", err)
			break
		}

		pbs := &primBlockSettings{
			int64(primitiveBlock.GetGranularity()),
			primitiveBlock.GetLatOffset(),
			primitiveBlock.GetLonOffset(),
			0.000000001} // Stolen from imposm3

		strTable := primitiveBlock.GetStringtable()
		primgroup := primitiveBlock.GetPrimitivegroup()

		wg.Add(1)

		for _, group := range primgroup {

			// var decodedN []*MyNode
			if len(group.GetDense().GetId()) != 0 {
				// wg.Add(1)
				go decodeDenseNodes(group, strTable, pbs, &wg)
				// allnodes = append(allnodes, decoded...)
				c_dense += 1
			}

			// removed because not working.
			// TODO replace by channels
			// for _, node := range decodedN {
			// 	largeMapNode[node.Id] = *node
			// 	if len(node.Tags) != 0 {
			// 		counter_tagged += 1
			// 	} else {
			// 		counter_simple += 1
			// 	}
			// }

			// TODO implement; but unessary for Geofabrik exports.
			if len(group.GetNodes()) != 0 {
				// wg.Add(1)
				c_node += 1
			}
			// var decodedW []*MyWay
			if len(group.GetWays()) != 0 {
				// wg.Add(1)
				go DecodeWays(group, strTable, &wg)
				// allWays = append(allWays, decoded...)
				c_way += 1
			}

			// removed because not working.
			// TODO replace by channels
			// for _, way := range decodedW {
			// 	if way.Type == "LineString" {
			// 		counter_LineString += 1
			// 	} else {
			// 		counter_Polygon += 1
			// 	}
			// }

			// for _, way := range decodedW {
			// 	largeMapWays[way.Id] = *way
			// }

			if len(group.GetRelations()) != 0 {
				// wg.Add(1)
				_ = DecodeRelations(group, strTable, &wg)
				c_relation += 1
			}
		}
		i += 1
	}
	wg.Wait()
	// Todo fix this use wait group..
	// time.Sleep(5 * time.Second)

	fmt.Println("Anzahl Primitive Blocks:")
	fmt.Printf("\t"+
		"DenseNodes: %v (Features -> total: %v, tagged: %v, simple: %v)\n\t"+
		"Nodes: %v\n\t"+
		"Ways: %v (Features -> total: %v, LineStrings: %v, Polygons: %v)\n\t"+
		"Relations: %v\n\t"+
		"Summe: %v\n",
		c_dense, (counter_tagged + counter_simple), counter_tagged, counter_simple,
		c_node,
		c_way, (counter_LineString + counter_Polygon), counter_LineString, counter_Polygon,
		c_relation,
		i)
	// TODO this is buggy somehow.
	fmt.Println("Anzahl nicht korrekt gelesener LineStrings:", FailCnt, "(Currently means way with no tags.)")

	// log.Println("Processing Nodes, Ways & Relations took ", time.Since(start))
	return nil
}

func extractBlob(blobHeader *pb.BlobHeader, file *os.File) ([]byte, error) {
	blob := pb.Blob{}

	blobdata := make([]byte, blobHeader.GetDatasize())
	io.ReadFull(file, blobdata)

	err := proto.Unmarshal(blobdata, &blob)
	if err != nil {
		return []byte{}, fmt.Errorf("UnmarshalBlob error, %v", err)
	}

	b := bytes.NewReader(blob.GetZlibData())
	r, err := zlib.NewReader(b)
	if err != nil {
		return []byte{}, fmt.Errorf("new Reader Error %v", err)
	}
	defer r.Close()

	builder := new(strings.Builder)
	io.Copy(builder, r)

	// TODO
	return []byte(builder.String()), nil
}

func extractBlobHeader(file *os.File) (*pb.BlobHeader, error) {
	var size int32
	err := binary.Read(file, binary.BigEndian, &size)
	if err != nil {
		log.Printf("\n %d, %T \n", err, err)
		return &pb.BlobHeader{}, fmt.Errorf("%v reading header size", err)
	}
	// log.Println(size)
	data := make([]byte, size)
	n, err := io.ReadFull(file, data)
	if err != nil {
		return &pb.BlobHeader{}, fmt.Errorf("ReadFull %v", err)
	}
	if n != int(size) {
		return &pb.BlobHeader{}, fmt.Errorf("reading blob header, only got %d bytes instead of %d", n, size)
	}

	blobHeader := pb.BlobHeader{}

	err = proto.Unmarshal(data, &blobHeader)
	if err != nil {
		return &pb.BlobHeader{}, fmt.Errorf("UnmarshalBlobHeader error, %v", err)
	}

	return &blobHeader, nil
}
