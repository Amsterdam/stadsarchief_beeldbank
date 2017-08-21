package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jinzhu/gorm"
	"github.com/kelseyhightower/envconfig"
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
	Identifier           string      `xml:"identifier"`
	Source               string      `xml:"source"`
	Type                 string      `xml:"type"`
	Title                string      `xml:"title"`
	GeoName              string      `xml:"subject"`
	Creator              string      `xml:"creator"`
	ParameterList        []Parameter `xml:"parameter"`
	Provenance           string      `xml:"provenance"`
	Rights               string      `xml:"rights"`
	DateText             string      `xml:"date"`
	Description          string      `xml:"description"`
	FileName             string
	DateFrom             string
	DateTo               string
	Levering             bool
	Leveringsvoorwaarden string
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
	locatieImageChan  chan *[]string
	metaImageChan     chan *[]string
	imagelocationChan chan *[]string
	metaImageColumns  []string
	locationColumns   []string
	Config            ConfigSpec
)

type ConfigSpec struct {
	Debug    bool   `default:"false"`
	Port     int    `default:"5432"`
	User     string `default:"beeldbank"`
	Password string `default:"insecure"`
	Database string `default:"beeldbank"`
	Host     string `default:"database"`
	DataPath string `default:"/app/data"`
}

func init() {
	imageCount = 0
	duplicates = 0
	success = 0
	failed = 0
	imageIds = make(map[string]BeeldbankImageXML)
	// TODO make environment variable
	metaImageChan = make(chan *[]string, 3000)
	locatieImageChan = make(chan *[]string, 3000)
	metaImageColumns = []string{
		"image_id",
		"type",
		"source",
		"title",
		"creator",
		"provenance",
		"rights",
		//"leverings_voorwaarden",
		//"levering",
		"date_text",
		"description",
		//"date_from",
		//"date_to",
		//"adres",
	}
	locationColumns = []string{
		"image_id",
		"streetname",
		"number_from",
		"number_to",
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

//parseXMLNode parse single rdf / xml description of image
//detects duplicate ImageID's
func parseXMLNode(decoder *xml.Decoder, xmlNode *xml.StartElement, sourcefile *string) {

	var bbImage BeeldbankImageXML
	var id string

	err := decoder.DecodeElement(&bbImage, xmlNode)
	if err != nil {
		panic(err.Error())
	}

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

func parseDateRange(dates string, image *BeeldbankImageXML) {
	sdates := strings.Split(dates, "-")

	if len(sdates) == 2 {
		image.DateFrom = sdates[0]
		image.DateTo = sdates[1]
	} else if len(dates) == 1 {
		image.DateFrom = sdates[1]
		image.DateTo = sdates[1]
	}
}

//sendMetaImageInChannel as string array
func sendMetaImageInChannel(image *BeeldbankImageXML) {

	imageinfo := []string{
		image.Identifier,  //image_id
		image.Type,        //type
		image.Source,      //source
		image.Title,       //title
		image.Creator,     //creator
		image.Provenance,  //provenance
		image.Rights,      //rights
		image.DateText,    //date_text
		image.Description, //description
	}

	for _, param := range image.ParameterList {
		//log.Println(param)
		switch paramName := param.Name; paramName {
		case "datering":
			//log.Println("datering")
			dates := param.Value
			parseDateRange(dates, image)
		case "levering":
			//log.Println("levering")
			if param.Value == "ja" {
				image.Levering = true
			} else {
				image.Levering = false
			}
		case "geografische naam":

			locatie := []string{
				image.Identifier,
				param.Straatnaam,
				param.NumberFrom,
				param.NumberTo,
			}
			locatieImageChan <- &locatie

		}
	}

	metaImageChan <- &imageinfo

}

//parse one source xml file
func parseSingleXML(sourcefile string) {

	log.Println("Parsing:", sourcefile)

	xmlfile, err := os.Open(sourcefile)
	defer xmlfile.Close()

	//bar = NewProgressBar(csvfile)

	if err != nil {
		log.Println(err)
		panic(err.Error())
	}

	decoder := xml.NewDecoder(xmlfile)

	for {
		// Read tokens from the XML document in a stream.
		token, err := decoder.Token()

		if token == nil {
			break
		}

		if err != nil {
			panic(err.Error())
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

	files, err := filepath.Glob(fmt.Sprintf("%s/b2*.xml", Config.DataPath))

	if err != nil {
		panic(err)
	}

	if len(files) == 0 {
		log.Printf(Config.DataPath)
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
	close(locatieImageChan)
}

func logcounts() {
	log.Printf("Parsed Images: %d   duplicates %d ", imageCount, duplicates)
}

func startImport() {
	//prepare database
	Migrate()
	//open database table
	wg.Add(2)
	go StreamInTable("beeldbank_images", metaImageColumns, metaImageChan)
	go StreamInTable("image_locations", locationColumns, locatieImageChan)
	//stream xml into database
	importXMLbeeldbank()
	wg.Wait()

	err := DB.Close()
	if err != nil {
		panic(err.Error())
	}
}

func setUpEnvironment() {
	//parse environment variables
	err := envconfig.Process("xmlparser", &Config)
	if err != nil {
		log.Fatal(err.Error())
	}

	DB = DBConnect(ConnectStr())
}

func main() {
	setUpEnvironment()
	startImport()
}
