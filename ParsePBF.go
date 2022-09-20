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
	"time"

	"google.golang.org/protobuf/proto"

	pb "github.com/Timahawk/osm_file_decoder/proto"
	"github.com/jackc/pgx/v5"
)

type PrimBlogSettings struct {
	granularity int64
	latOffset   int64
	lonOffset   int64
	coordScale  float64
}

func LoopOverFile(file *os.File, conn *pgx.Conn) error {
	// BlobHeader und Blob kommen aus dem "fileformat.proto"
	// HeaderBlog und PrimitiveBlog kommen aus "osmformat.proto"

	// Das ist ein BlobHeader
	blobHeader, err := extractHeader(file)
	if err != nil {
		log.Fatalln("Error reading the first BlobHeader")
	}

	// Das  ist der Blob der den HeaderBlog enthält
	blob, err := extractBlob(blobHeader, file)
	if err != nil {
		log.Fatalln("Error reading the HeaderBlog")
	}
	headerBlog := pb.HeaderBlock{}
	err = proto.Unmarshal(blob, &headerBlog)
	if err != nil {
		log.Fatalf("UnmarshalBlob error, %v", err)
	}
	fmt.Println(
		"Bbox", headerBlog.GetBbox(),
		"RequiredFeatures", headerBlog.GetRequiredFeatures(),
		"OptinalFeatures", headerBlog.GetOptionalFeatures(),
		"WritingProgramm", headerBlog.GetWritingprogram())

	// Das ist ein BlobHeader
	blobHeader, err = extractHeader(file)
	if err != nil {
		log.Fatalln("Error reading BlobHeader2")
	}

	// Hier gehts quasi richtig los mit den Daten
	// Im ersten Blog sind aber nur DenseNodes
	// DenseNodes sind nicht ein Struct mit jeweils Lat Long ID undsowas,
	// Sondern eine Liste die dann zusammengebastelt werden muss.
	// Die Eigenschaften kommen dann aus dem String Tabel
	// Ist alles ein wenig umständlich gemacht...
	blog, err := extractBlob(blobHeader, file)
	if err != nil {
		log.Fatalln("Error reading Blob2")
	}

	primitiveBlog := pb.PrimitiveBlock{}
	err = proto.Unmarshal(blog, &primitiveBlog)
	if err != nil {
		log.Fatalf("UnmarshalBlob error, %v", err)
	}

	pbs := &PrimBlogSettings{
		int64(primitiveBlog.GetGranularity()),
		primitiveBlog.GetLatOffset(),
		primitiveBlog.GetLonOffset(),
		0.000000001}

	i := 0

	// testfile, _ := os.Create("testfile2.csv")
	// defer testfile.Close()
	// w := csv.NewWriter(testfile)

	// wayfile, _ := os.Create("wayfile.csv")
	// defer wayfile.Close()
	// waywriter := csv.NewWriter(wayfile)

	counter_simple := 0
	counter_tagged := 0

	largeMap := make(map[int64]Coord)

	c_dense, c_node, c_way, c_relation := 0, 0, 0, 0
	start := time.Now()
	// Ab hier wird dann solange über abwechselnd BlobHeader Blob geloopt bis error.
	// In den Blobs sind nur noch blogs!
	for {
		blobHeader, err := extractHeader(file)
		if err != nil {
			fmt.Printf("Error reading BlobHeader2 %v\n", err)
			break
		}

		blog, err := extractBlob(blobHeader, file)
		if err != nil {
			fmt.Printf("Error reading Blob2\n")
			break
		}

		primitiveBlog := pb.PrimitiveBlock{}
		err = proto.Unmarshal(blog, &primitiveBlog)
		if err != nil {
			fmt.Printf("UnmarshalBlob error, %v\n", err)
			break
		}

		strTable := primitiveBlog.GetStringtable()
		primgroup := primitiveBlog.GetPrimitivegroup()

		allnodes := []*MyNode{}
		allWays := []*MyWay{}

		for _, group := range primgroup {

			if len(group.GetDense().GetId()) != 0 {
				allnodes = append(allnodes, decodeDenseNodes(group, strTable, pbs, conn)...)
				c_dense += 1
			}

			for _, node := range allnodes {
				largeMap[node.id] = Coord{node.lat, node.lon}
			}

			// for key, value := range largeMap {
			// 	fmt.Println("LARGE MAP", key, value)
			// }

			if len(group.GetNodes()) != 0 {
				c_node += 1
			}
			if len(group.GetWays()) != 0 {
				allWays = append(allWays, DecodeWays(group, strTable, largeMap, conn)...)
				c_way += 1
			}
			if len(group.GetRelations()) != 0 {
				c_relation += 1
			}
		}

		// 	if len(node.tags) != 0 {
		// 		w.Write([]string{
		// 			fmt.Sprintf("%v", node.id),
		// 			fmt.Sprintf("%f", node.lat),
		// 			fmt.Sprintf("%f", node.lon),
		// 			fmt.Sprintf("%v", node.tags)})
		// 		counter_tagged += 1
		// 	} else {
		// 		counter_simple += 1
		// 	}

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
	log.Println("Processing Nodes, Ways & Relations took ", time.Since(start))
	fmt.Println("dense", c_dense, "nodes", c_node, "way", c_way, "relation", c_relation)
	fmt.Println("Anzahl PrimitiveGroups:", i)
	fmt.Println("Anzahl einfacher Features:", counter_simple)
	fmt.Println("Anzahl taggeder Features:", counter_tagged)
	fmt.Println("Anzahl nicht korrekt gelesener LineStrings:", failCnt)
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

func extractHeader(file *os.File) (*pb.BlobHeader, error) {
	var size int32
	err := binary.Read(file, binary.BigEndian, &size)
	// if err == "EOF" {
	// 	return &pb.BlobHeader{}, nil
	// }
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
