package main

import (
	"fmt"
	"sync"

	pb "github.com/Timahawk/osm_file_decoder/proto"
)

type MyRelation struct {
	Id       int64
	Tags     map[string]string
	Refs     []int64  // memids I think
	RefsType []string // MemberTypes types
	RolesSid []int32  // No idea what this is for...
	Coords   []Coord
	Type     string
}

func DecodeRelations(pg *pb.PrimitiveGroup, st *pb.StringTable, wg *sync.WaitGroup) []*MyRelation {
	defer wg.Done()
	myrelations := []*MyRelation{}

	strtable := st.GetS()

	relations := pg.GetRelations()

	// var sb strings.Builder
	// var sbp strings.Builder

	// sb.WriteString("INSERT INTO Lines VALUES ")
	// sbp.WriteString("INSERT INTO Polygons VALUES ")
	// fmt.Println(sb.Len())

	for _, rela := range relations {

		mr := &MyRelation{rela.GetId(), map[string]string{}, []int64{}, []string{}, []int32{}, []Coord{}, ""}

		keys := rela.GetKeys()
		values := rela.GetVals()

		for i, k := range keys {
			mr.Tags[string(strtable[k])] = string(strtable[values[i]])
		}

		if val, ok := mr.Tags["role"]; ok {
			//do something here
			fmt.Println(val)
		}

		refs := rela.GetMemids()
		ref_type := rela.GetTypes()

		delta_ref := int64(0)

		for i, k := range refs {
			delta_ref = k + delta_ref

			switch {
			case ref_type[i] == pb.Relation_NODE:
				mr.Refs = append(mr.Refs, delta_ref)
				mr.RefsType = append(mr.RefsType, "Point")
				coords, _ := largeMapNode.Get(fmt.Sprint(delta_ref))
				mr.Coords = append(mr.Coords, coords.Coords)

			case ref_type[i] == pb.Relation_WAY:
				mr.Refs = append(mr.Refs, delta_ref)
				mr.RefsType = append(mr.RefsType, "Way")
				coords, _ := largeMapWays.Get(fmt.Sprint(delta_ref))
				mr.Coords = append(mr.Coords, coords.Coords...)

				// TODO Coords

			case ref_type[i] == pb.Relation_RELATION:
				mr.Refs = append(mr.Refs, delta_ref)
				mr.RefsType = append(mr.RefsType, "Relation")
				// TODO Coords
			}
		}
		roles := rela.GetRolesSid()

		// some fancy replacement for looping and appending each individally.
		mr.RolesSid = append(mr.RolesSid, roles...)

		// 	if mw.Refs[0] == mw.Refs[len(mw.Refs)-1] {
		// 		mw.Type = "Polygon"
		// 	} else {
		// 		mw.Type = "LineString"
		// 	}

		myrelations = append(myrelations, mr)

		// 	str := mw.String()
		// 	if str == "" {
		// 		failCnt += 1
		// 		fmt.Println("Failcnt", failCnt)
		// 	}
		// 	if mw.Type == "LineString" {
		// 		sb.WriteString(str)
		// 	}
		// 	if mw.Type == "Polygon" {
		// 		sbp.WriteString(str)
		// 	}
		// }

		// if ToDB_LineString {
		// 	// Catch no Nodes to write.
		// 	if sb.Len() > 30 {
		// 		str := sb.String()[:len(sb.String())-2]
		// 		err := Insert(conn, str)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 	}

		// }

		// if ToDB_Polygons {
		// 	// Catch no Nodes to write.
		// 	if sbp.Len() > 30 {
		// 		str := sbp.String()[:len(sbp.String())-2]
		// 		err := Insert(conn, str)
		// 		if err != nil {
		// 			log.Fatal(err)
		// 		}
		// 	}
		// }
	}
	return myrelations
}
