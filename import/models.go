package main

import "time"

var (
	imageTable				 string
	locationTable			 string
	fileTable				 string
	beeldbankImageColumns    []string
	imageLocationColumns     []string
	imageFileLocationColumns []string
)

func init() {
	imageTable = "beeldbank_images"
	locationTable = "image_locations"
	fileTable = "image_file_locations"
	beeldbankImageColumns = []string{
		"image_id",
		"type",
		"source",
		"title",
		"creator",
		"provenance",
		"rights",
		"date_text",
		"description",
		"date_from",
		"date_to",
		"levering",
		"leverings_voorwaarden",
	}
	imageLocationColumns = []string{
		"image_id",
		"streetname",
		"number_from",
		"number_to",
	}
	imageFileLocationColumns = []string{
		"image_id",
		"objectstore_path",
	}
}

//	LocatieModel of and Image. One image can have a multiple location indications
type ImageLocation struct {
	ID         uint `gorm:"primary_key"`
	ImageID    string
	Streetname string
	NumberFrom int
	NumberTo   int
	CreatedAt  time.Time
	Geom       GeoPoint `sql:"type:geometry(Geometry,4326)"`
}

//	BeeldbankImageModel database model
type BeeldbankImage struct {
	ID                   uint   `gorm:"primary_key"`
	ImageID              string `gorm:"unique_index"`
	Source               string
	Type                 string
	Title                string
	Creator              string
	LeveringsVoorwaarden string
	Levering             bool
	Provenance           string
	Rights               string
	CreatedAt            time.Time
	DateText             string
	DateFrom             string
	DateTo               string
	Description          string
}

//	ImageFileLocation database model
type ImageFileLocation struct {
	ID                   uint   `gorm:"primary_key"`
	ImageID              string `gorm:"unique_index"`
	ObjectstorePath      string
}

type ImageParameters struct {
	ID      uint   `gorm:"primary_key"`
	ImageID string `gorm:"unique_index"`
}
