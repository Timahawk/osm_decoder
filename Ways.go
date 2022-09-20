package main

import pb "github.com/Timahawk/osm_file_decoder/proto"

type Coord struct {
	lat float64
	lon float64
}
type MyWay struct {
	Id     int64
	Tags   map[string]string
	Coords []Coord
}

func DecodeWays(pg *pb.PrimitiveGroup, st *pb.StringTable) []*MyWay {

	myways := []*MyWay{}

	strtable := st.GetS()
	delta_id := int64(0)
	ways := pg.GetWays()

	for _, way := range ways {

		delta_id += way.GetId()
		mw := &MyWay{delta_id, map[string]string{}, []Coord{}}

		keys := way.GetKeys()
		values := way.GetVals()
		// refs := way.GetRefs()

		for i, k := range keys {
			mw.Tags[string(strtable[k])] = string(strtable[values[i]])
		}

		myways = append(myways, mw)
	}

	return myways
}
