#!/bin/bash

set -e
set -u

DIR="$(dirname $0)"

echo $0

echo "Do we have OS password?"
echo $BEELDBANK_OBJECTSTORE_PASSWORD
#


dc() {
	#docker-compose -p bb -f ../docker-compose.yml $*;
	docker-compose $*;
}

# so we can delete named volumes
#dc stop
#dc rm -f -v

dc pull

dc up -d database

dc exec database update-table.sh bag bag_nummeraanduiding public beeldbank
dc exec database update-table.sh bag bag_verblijfsobject public beeldbank

dc run -T importer ./import-xml.sh
#python objectstore.py
#dc exec importer ./xmlparser




