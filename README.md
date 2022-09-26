# osm_decoder

A new implementation an OSM.pbf File Reader. 
There is particular reason, as why this is needed or usefull, I simply wanted to try it out...

It uses [orb](https://pkg.go.dev/github.com/paulmach/orb) and [geojson](https://pkg.go.dev/github.com/paulmach/orb/geojson) for the Features.

## Features Implemented:
 - Reads the PBF File 
 - Writes the Nodes/ Points into Postgres
 - Write Ways/Lines into Postgres
 - Writes Ways/Polygons into Postgres

## Disclaimer:
 - Right now, the writing of Ways is **NOT** correct. (Because of Mulitpolygons with holes for example.)
 - This programm is **NOT** ready for production. Very likely never will be.
 - Use at your own risk. 
 - Very likely to break for large files and limitied memory. (Not tested.)
## Next:
 - Fix false Geometry with Lines & Polygons
 - Implement Relations.
 - Implement a Proof of Work (only Points) to Output MapboxVectorTiles Protobuf. Similar to [Tilemaker](https://github.com/systemed/tilemaker). 