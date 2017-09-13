# stadsarchief_beeldbank
Indexerenen / Zoekfunctionaliteit van historische beelden


Import
------

Importeren van xml / rdf bestanden van het stadsarchief.

environment varibales for import.

    BEELDBANK_DEBUG = False
    BEELDBANK_PORT = 5432
    BEELDBANK_HOST = "database"
    BEELDBANK_USER = "beeldbank"
    BEELDBANK_DATA_PATH = "/app/data"
    BEELDBANK_DATABASE = "beeldbank"
    BEELDBANK_PASSWORD = "insecure"

To download the latest xml / rdf(ish) files.

	  # make sure that the BEELDBANK_OBJECTSTORE_PASSWORD is available:
      docker-compose build importer
      docker-compose run importer python objectstore.py

Start a database

      docker-compose up -d database

Run the go xml importer.

      docker-compose run importer xmlparser

Image Geo-location
-----------

We want Images to be plotted on a map.
With blunt Bag name mathcing 40% of images can
be fixed by just looking directly to bag.

TODO use advanced search engine (variation of HR code)
TODO use kadaster historical data?

Current process

1) Download latest bag data

    docker-compose exec database update-table.sh bag bag_nummeraanduiding public beeldbank
    docker-compose exec database update-table.sh bag bag_verblijfsobject public beeldbank

2) execute sql plan

    docker-compose run importer python bag_sql_recipes.py


API
---

todo..

