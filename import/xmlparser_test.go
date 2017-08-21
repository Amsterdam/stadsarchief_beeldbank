package main

import (
	"testing"
	"time"
)

func init() {
	setUpEnvironment()
	Config.DataPath = "."
	Config.Debug = true
	startImport()
	//importXMLbeeldbank()
}

func TestParsing(t *testing.T) {
	if len(imageIds) != 2 {
		t.Error("Parsing failed..")
	}
}

func TestParsingInDB(t *testing.T) {

	time.Sleep(1000 * time.Millisecond)
	var count int64
	db := DBConnect(ConnectStr())
	db.Model(&BeeldbankImage{}).Count(&count)
	if count != 2 {
		t.Error("Datbase count incorrect..", count)
	}
}
