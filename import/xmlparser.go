//	Parse the beeldbank xml files and load the data / attributes in database.
//	A seperate location indication table is created which is later used to determine exect geo-loaction
//	of images in the "beeldbank" image archive.
package main

import (
	"encoding/xml"
	"log"
	"os"
	"strings"
)

//	Parameter information on images extra variables
type Parameter struct {
	Name       string `xml:"name,attr"`
	Value      string `xml:",chardata"`
	Straatnaam string `xml:"name"`
	NumberFrom string `xml:"number_from"`
	NumberTo   string `xml:"number_to"`
}

//	BeeldbankImageXML is XML image mapping
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
	//	total found images
	imageCount int
	duplicates int
	//	source of beeldbank xml files
	dataPath 		 string

)

func init() {
	imageCount = 0
	duplicates = 0
	imageIds = make(map[string]BeeldbankImageXML)
	dataPath = "/app/data"

	// TODO make environment variable
}

//	logdupes prints two xml image enries side by side to compare attributes
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

//	parseXMLNodeToChannel parse single rdf / xml description of image
//	detects duplicate ImageID's and puts results on channels
func parseXMLNodeToChannel(decoder *xml.Decoder, xmlNode *xml.StartElement, sourcefile *string,
	metaImageChan chan *[]string, locationChan chan *[]string) {

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
		metaImage, locations := parseImageXML(&bbImage)
		metaImageChan <- &metaImage
		for _, location := range locations {
			locationChan <- &location
		}
	}
}

func parseDateRange(dates string, image []string) {
	sdates := strings.Split(dates, "-")

	if len(sdates) == 2 {
		// date_from , date_to
		image[9] = sdates[0]
		image[10] = sdates[1]
	} else if len(dates) == 1 {
		// date_from = date_to
		image[9] = sdates[1]
		image[10] = sdates[1]
	}
}

//	parseImageXML as string array
func parseImageXML(image *BeeldbankImageXML) ([]string, [][]string) {
	var locations [][]string

	metaImage := []string{
		image.Identifier,  //0  image_id
		image.Type,        //1  type
		image.Source,      //2  source
		image.Title,       //3  title
		image.Creator,     //4  creator
		image.Provenance,  //5  provenance
		image.Rights,      //6  rights
		image.DateText,    //7  date_text
		image.Description, //8  description
		"",                //9  date_from
		"",                //10 date_to
		"",                //11 levering
		"",                //12 leveringsvoorwaarden
	}

	for _, param := range image.ParameterList {
		switch paramName := param.Name; paramName {
		case "datering":
			dates := param.Value
			parseDateRange(dates, metaImage)

		case "levering":
			if param.Value == "ja" {
				metaImage[11] = "1"
				image.Levering = true
			} else {
				metaImage[11] = "0"
				image.Levering = false
			}

		case "geografische naam":
			locations = append(locations, []string{
				image.Identifier,
				param.Straatnaam,
				param.NumberFrom,
				param.NumberTo,
			})
		}
	}

	return metaImage, locations
}

//	parse one source xml file
func parseSingleXMLFileTo(sourcefile string, metaImageChan chan *[]string, locationChan chan *[]string) {
	log.Println("Parsing:", sourcefile)

	xmlfile, err := os.Open(sourcefile)
	defer xmlfile.Close()

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
			// If we just read a StartElement token and its name is "rdf:Description"
			if xmlNode.Name.Local == "Description" {
				imageCount++
				// decode a whole chunk of following XML into the variable bbImage which is a BeeldbankImageXML
				parseXMLNodeToChannel(decoder, &xmlNode, &sourcefile, metaImageChan, locationChan)
			}
		}
	}

	//	prints some stats.
	log.Printf("Parsed Images: %d   duplicates %d ", imageCount, duplicates)
}

