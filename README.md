# stadsarchief_beeldbank
Indexerenen / Zoekfunctionaliteit van historische beelden


Import
------

Importeren van xml / rdf bestanden van het stadsarchief.

environment varibales for import.

    XMLPARSER_DEBUG = False
    XMLPARSER_PORT = 5432
    XMLPARSER_HOST = "database"
    XMLPARSER_USER = "beeldbank"
    XMLPARSER_DATA_PATH = "/app/data"
    XMLPARSER_DATABASE = "beeldbank"
    XMLPARSER_PASSWORD = "insecure"




To download the latest xml / rdf(ish) files.

      python objectstore.py

Start a database

      docker-compose up database

Run the go xml importer.

    go get
    go build
    ./xmlparser


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

    python bag_sql_recepes.py


API
---

todo..

