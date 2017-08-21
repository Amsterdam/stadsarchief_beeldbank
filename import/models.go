package main

import "time"

//LocatieModel of and Image. One image can have a multiple
//location indications
type ImageLocation struct {
	ID             uint `gorm:"primary_key"`
	ImageID        string
	AdresIndicatie string
	HuisnummerFrom int
	HuisnummerTo   int
	CreatedAt      time.Time
	Geom           GeoPoint `sql:"type:geometry(Geometry,4326)"`
}

//BeeldbankImageModel database model
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

type ImageParameters struct {
	ID      uint   `gorm:"primary_key"`
	ImageID string `gorm:"unique_index"`
}
