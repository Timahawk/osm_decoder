package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"google.golang.org/protobuf/proto"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/jackc/pgx/v5"
)

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
func LoopOverFile(file *os.File, conn *pgx.Conn) error {

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

	largeMap := make(map[int64]Coord)

	c_dense, c_node, c_way, c_relation := 0, 0, 0, 0

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
		if err == errors.New("EOF") {
			log.Println("Reached end of file after reading", i, "Blocks.")
			break

		}
		if err != nil {
			fmt.Printf("Error reading BlobHeader2 %v\n", err)
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

		allnodes := []*MyNode{}
		allWays := []*MyWay{}

		for _, group := range primgroup {

			if len(group.GetDense().GetId()) != 0 {
				decoded := decodeDenseNodes(group, strTable, pbs, conn)
				allnodes = append(allnodes, decoded...)
				c_dense += 1
			}

			for _, node := range allnodes {
				largeMap[node.Id] = Coord{node.Lat, node.Lon}
				if len(node.tags) != 0 {
					counter_tagged += 1
				} else {
					counter_simple += 1
				}
			}

			// for key, value := range largeMap {
			// 	fmt.Println("LARGE MAP", key, value)
			// }

			if len(group.GetNodes()) != 0 {
				c_node += 1
			}
			if len(group.GetWays()) != 0 {
				decoded := DecodeWays(group, strTable, largeMap, conn)
				allWays = append(allWays, decoded...)
				c_way += 1
			}

			for _, way := range allWays {
				if way.Type == "LineString" {
					counter_LineString += 1
				} else {
					counter_Polygon += 1
				}
			}

			if len(group.GetRelations()) != 0 {
				c_relation += 1
			}
		}

		// 	// err = db.Update(func(tx *badger.Txn) error {
		// 	// 	key := make([]byte, 8)
		// 	// 	binary.BigEndian.PutUint64(key, uint64(node.id))

		// 	// 	value := make([]byte, 8)
		// 	// 	binary.BigEndian.PutUint64(value, math.Float64bits(node.lat))

		// 	// 	tx.Set(key, value)
		// 	// 	// tx.Commit()
		// 	// 	return nil
		// 	// })
		// 	// if err != nil {
		// 	// 	log.Fatalln(err)
		// 	// }
		// }
		// w.Flush()

		// for _, way := range allWays {

		// fmt.Println(way.String())

		// 	waywriter.Write([]string{
		// 		fmt.Sprintf("%v", way.Id),
		// 		fmt.Sprintf("%v", way.Tags),
		// 		fmt.Sprintf("%v", way.Refs),
		// 		fmt.Sprintf("%v", way.Coords)})

		// }
		// waywriter.Flush()

		i += 1
	}

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
	fmt.Println("Anzahl nicht korrekt gelesener LineStrings:", failCnt)

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
	if err == errors.New("EOF") {
		log.Println("Reached end of file.")
		return &pb.BlobHeader{}, err
	}
	if err != nil {
		return &pb.BlobHeader{}, fmt.Errorf("%v reading header size", err)
	}

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
