package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/jinzhu/gorm"
)

//Parameter information on images extra variables
type Parameter struct {
	Name       string `xml:"name,attr"`
	Value      string `xml:",chardata"`
	Straatnaam string `xml:"name"`
	NumberFrom string `xml:"number_from"`
	NumberTo   string `xml:"number_to"`
}

//BeeldbankImageXML is XML image mapping
type BeeldbankImageXML struct {
	Identifier    string      `xml:"identifier"`
	Source        string      `xml:"source"`
	Type          string      `xml:"type"`
	Title         string      `xml:"title"`
	GeoName       string      `xml:"subject"`
	Creator       string      `xml:"creator"`
	ParameterList []Parameter `xml:"parameter"`
	FileName      string
	//Number,
	//Description,
	//Rights,
	//Date,
	//Dataclean,
	//Levering      string
	//Leveringsvoorwaarden,
}

var (
	imageIds map[string]BeeldbankImageXML
	// total found images
	imageCount int
	duplicates int
	success    int
	failed     int
	wg         sync.WaitGroup
	DB         *gorm.DB
	// source of beeldbank xml files
	sourceXMLdir        string
	metaImageChan       chan *[]string
	metaImageColumns    []string
	locationChan        chan *[]string
	locationChanColumns []string
)

func init() {
	imageCount = 0
	duplicates = 0
	success = 0
	failed = 0
	imageIds = make(map[string]BeeldbankImageXML)
	// TODO make environment variable
	sourceXMLdir = "/app/data"
	metaImageChan = make(chan *[]string, 3000)
	metaImageColumns = []string{
		"image_id",
		//"source",
		"type",
		//"adres",
	}
}

func logdupes(i1 BeeldbankImageXML, i2 BeeldbankImageXML) {

	log.Printf(`
id	%-15s  %15s
type	%-15s  %15s
title	%-15s  %15s
xml	%-15s  %15s
geo	%-15s  %15s
creator %-15s  %15s
	`, i1.Identifier, i2.Identifier,
		i1.Type, i2.Type,
		i1.Title, i1.Title,
		i1.FileName, i2.FileName,
		i1.GeoName, i2.GeoName,
		i1.Creator, i2.Creator,
	)
}

//parse single rdf / xml description of image
func parseXMLNode(decoder *xml.Decoder, xmlNode *xml.StartElement, sourcefile *string) {

	var bbImage BeeldbankImageXML
	var id string

	decoder.DecodeElement(&bbImage, xmlNode)

	id = bbImage.Identifier
	bbImage.FileName = *sourcefile

	if _, ok := imageIds[id]; ok {
		log.Println("DUPLICATES FOUND! : ", id)
		logdupes(imageIds[id], bbImage)
		duplicates++
	} else {
		imageIds[id] = bbImage
		sendMetaImageInChannel(&bbImage)
	}

}

//sendMetaImageInChannel as string array
func sendMetaImageInChannel(image *BeeldbankImageXML) {

	array := []string{
		image.Identifier, //image_id
		image.Type,       //type
	}

	metaImageChan <- &array
}

//parse one source xml file
func parseSingleXML(sourcefile string) {

	log.Println("Parsing:", sourcefile)

	xmlfile, err := os.Open(sourcefile)
	defer xmlfile.Close()

	//bar = NewProgressBar(csvfile)

	if err != nil {
		log.Println(err)
		panic(err)
	}

	decoder := xml.NewDecoder(xmlfile)

	for {
		// Read tokens from the XML document in a stream.
		token, err := decoder.Token()

		if token == nil {
			break
		}

		if err != nil {
			panic(err)
		}

		// Inspect the type of the token just read.
		switch xmlNode := token.(type) {

		case xml.StartElement:
			// If we just read a StartElement token
			// ...and its name is "rdf:Description"
			if xmlNode.Name.Local == "Description" {
				imageCount++
				// decode a whole chunk of following XML into the
				// variable bbImage which is a BeeldbankImageXML
				parseXMLNode(decoder, &xmlNode, &sourcefile)
			}
		}
	}
	//prints some stats.
	logcounts()
}

func findXMLFiles() []string {

	files, err := filepath.Glob(fmt.Sprintf("%s/b2*.xml", sourceXMLdir))

	if err != nil {
		panic(err)
	}

	if len(files) == 0 {
		log.Printf(sourceXMLdir)
		panic(errors.New("Missing XML files"))
	}

	return files
}

func importXMLbeeldbank() {

	files := findXMLFiles()

	for _, file := range files {
		parseSingleXML(file)
	}

	close(metaImageChan)
}

func logcounts() {
	log.Printf("Parsed Images: %d   duplicates %d ", imageCount, duplicates)
}

func main() {

	DBConnect(ConnectStr())

	Migrate()
	wg.Add(1)
	go StreamInTable("beeldbank_images", metaImageColumns, metaImageChan)
	importXMLbeeldbank()
	wg.Wait()
	DB.Close()
	logcounts()
}
