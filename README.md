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


API
---

todo..

