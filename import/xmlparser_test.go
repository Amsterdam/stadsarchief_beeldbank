package main

import (
	"testing"
	"time"
)

func init() {
	DBConfig.Debug = true
	DB := DBConnect("beeldbank")
	runImport(DB)
}

func TestParsing(t *testing.T) {
	if len(imageIds) != 2 {
		t.Error("Parsing failed..")
	}
}

func TestParsingInDB(t *testing.T) {
	time.Sleep(1000 * time.Millisecond)
	var count int64
	db := DBConnect("beeldbank")
	db.Model(&BeeldbankImage{}).Count(&count)
	if count != 2 {
		t.Error("Datbase count incorrect..", count)
	}
}
