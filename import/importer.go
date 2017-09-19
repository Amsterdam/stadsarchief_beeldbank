package main

import (
	"fmt"
	"log"
	"os"
	"bufio"
	"strings"
	"github.com/jinzhu/gorm"
	"sync"
	"path/filepath"
	"errors"
)

func findXMLFiles() []string {
	files, err := filepath.Glob(fmt.Sprintf("%s/b2*.xml", dataPath))

	if err != nil {
		panic(err)
	}

	if len(files) == 0 {
		log.Printf(dataPath)
		panic(errors.New("Missing XML files"))
	}

	return files
}

func queueFileListTo(imagefileChan chan *[]string) {
	imagelist := fmt.Sprintf("%s/image_list.txt", dataPath)
	log.Println("Parsing: ", imagelist)

	if file, err := os.Open(imagelist); err == nil {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			path := strings.Split(line, "/")
			filename := path[len(path)-1]

			if strings.Contains(filename, ".") {
				image_id := strings.Split(filename, ".")[0]
				imagefile := []string{
					image_id,
					line,
				}

				imagefileChan <- &imagefile
			}
		}

		if err = scanner.Err(); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}

func runImport(DB *gorm.DB) {
	var wg sync.WaitGroup

	//	stream xml into database
	imageChan := make(chan *[]string, 3000)
	locationChan := make(chan *[]string, 3000)

	wg.Add(2)
	go StreamInTable(imageTable, beeldbankImageColumns, imageChan, DB, &wg)
	go StreamInTable(locationTable, imageLocationColumns, locationChan, DB, &wg)

	files := findXMLFiles()
	for _, file := range files {
		parseSingleXMLFileTo(file, imageChan, locationChan)
	}

	close(locationChan)
	close(imageChan)

	//	stream file entries into database
	imagefileChan := make(chan *[]string, 3000)

	wg.Add(1)
	go StreamInTable(fileTable, imageFileLocationColumns, imagefileChan, DB, &wg)

	queueFileListTo(imagefileChan)
	close(imagefileChan)

	wg.Wait()
}

func main() {
	//	get databaseconnection
	DB := DBConnect("beeldbank")
	defer DBClose(DB)

	//  prepare database
	Migrate(DB)

	//  start import to database
	runImport(DB)
}

