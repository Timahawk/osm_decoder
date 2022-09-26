package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// func toDB(pool *pgxpool.Pool) {}

func toDB(pool *pgxpool.Pool) {

	var sb, sbb strings.Builder
	sb.WriteString("Insert into Points (id,tags,geom) VALUES ")

	valuesNodes := largeMapNode.IterBuffered()
	for tuple := range valuesNodes {
		if len(tuple.Val.Feature.Properties) != 0 {

			sb.WriteString(tuple.Val.SQLString() + ",")
		}
	}
	// TODO commit more often.
	str := sb.String()
	err := Insert(pool, str[:len(str)-1])
	if err != nil {
		log.Fatalln(err)
	}

	sb.Reset()
	sb.WriteString("Insert into Lines (id,tags,geom) VALUES")
	sbb.WriteString("Insert into Polygons (id,tags,geom) VALUES")

	valuesLineString := largeMapWays.IterBuffered()
	for tuple := range valuesLineString {
		if tuple.Val.Feature.Geometry.GeoJSONType() == "LineString" {
			if len(tuple.Val.Feature.Properties) != 0 {
				sb.WriteString(tuple.Val.SQLString() + ",")
			}
		} else {
			if len(tuple.Val.Feature.Properties) != 0 {
				sbb.WriteString(tuple.Val.SQLString() + ",")
			}
		}

	}
	// TODO commit more often.
	str = sb.String()
	err = Insert(pool, str[:len(str)-1])
	if err != nil {
		log.Fatalln(err)
	}
	str = sbb.String()
	err = Insert(pool, str[:len(str)-1])
	if err != nil {
		log.Fatalln(err)
	}

	// sb.Reset()
	// sb.WriteString("Insert into Polygons (id,tags,geom) VALUES")
	// valuesLinePolygons := largeMapPolygon.IterBuffered()
	// for tuple := range valuesLinePolygons {
	// 	if len(tuple.Val.Tags) != 0 {
	// 		sb.WriteString(tuple.Val.SQLString() + ",")
	// 	}
	// }
	// // TODO commit more often.
	// str = sb.String()
	// err = Insert(pool, str[:len(str)-1])
	// if err != nil {
	// 	log.Fatalln(err)
	// }
}

func Insert(pool *pgxpool.Pool, str string) error {

	_, err := pool.Exec(context.Background(), str)
	if err != nil {
		// return fmt.Errorf("exec: %d,\n %s ", err, str)
		return fmt.Errorf("exec:, %d,\n %s,\n\n %s", err, str[:50000], str[len(str)-500:])
	}

	return nil
}
