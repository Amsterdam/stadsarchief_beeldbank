package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/lib/pq"
)

type Locatie struct {
	ID              uint `gorm:"primary_key"`
	ImageID         string
	AdresIndicatie  string
	Huisnummer_from int
	Huisnummer_to   int
	CreatedAt       time.Time

	Geom GeoPoint `sql:"type:geometry(Geometry,4326)"`
}

type BeelbankImage struct {
	ID                   uint   `gorm:"primary_key"`
	ImageID              string `gorm:"unique_index"`
	Source               string
	Type                 string
	Creator              string
	LeveringsVoorwaarden string
	Levering             bool
	CreatedAt            time.Time
	DateringText         string
	DateFrom             string
	DateTo               string
	Geom                 GeoPoint `sql:"type:geometry(Geometry,4326)"`
}

//ConnectStr create string to connect to database
func ConnectStr() string {

	otherParams := "sslmode=disable connect_timeout=5"
	return fmt.Sprintf(
		"user=%s dbname=%s password='%s' host=%s port=%s %s",
		"beeldbank",
		"beeldbank",
		"insecure",
		"database",
		"5432",
		otherParams,
	)
}

//setup a database connection
func dbConnect(connectStr string) (db *gorm.DB, err error) {
	//db, err := sql.Open("postgres", connectStr)
	db, err = gorm.Open("postgres", connectStr)

	if err != nil {
		return
	}

	return db, nil

}

func Migrate() {
	log.Printf("Create db tables..")

	db, err := dbConnect(ConnectStr())

	if err != nil {
		panic(err)
	}

	db.DropTableIfExists(&Locatie{}, &BeelbankImage{})
	//Db = db
	db.AutoMigrate(&Locatie{}, &BeelbankImage{})

	defer db.Close()
}

//SQLImport import structure
type SQLImport struct {
	txn  *sql.Tx
	stmt *sql.Stmt
}

//AddRow Add a single row of data to the database
func (i *SQLImport) AddRow(columns ...interface{}) error {
	_, err := i.stmt.Exec(columns...)
	return err
}

//Commit the import to database
func (i *SQLImport) Commit() error {

	_, err := i.stmt.Exec()
	if err != nil {
		return err
	}

	// Statement might already be closed
	// therefore ignore errors
	_ = i.stmt.Close()

	return i.txn.Commit()
}

//NewImport setup a new import struct
func NewImport(db *sql.DB, schema string, tableName string, columns []string) (*SQLImport, error) {

	txn, err := db.Begin()

	if err != nil {
		return nil, err
	}

	stmt, err := txn.Prepare(pq.CopyInSchema(schema, tableName, columns...))
	if err != nil {
		return nil, err
	}

	return &SQLImport{txn, stmt}, nil
}

func normalizeRow(record *[]string) ([]interface{}, error) {

	cols := make([]interface{}, len(*record))

	for i, field := range *record {

		cols[i] = field

	}

	return cols, nil
}

//print columns we try to put in database
func printCols(cols []interface{}) {
	log.Println("\ncolumns:")
	for i, field := range cols {
		log.Printf("%2d %32s", i, field)
	}
}

func streamInTable(tablename string, columns []string, rows <-chan *[]string) error {

	db, err := dbConnect(ConnectStr())

	if err != nil {
		panic(err)
	}

	cdb := db.CommonDB().(*sql.DB)

	pgTable, err := NewImport(cdb, "public", tablename, columns)

	if err != nil {
		panic(err)
	}

	for row := range rows {

		cols, err := normalizeRow(row)

		if err != nil {
			log.Println(err)
			failed++
			continue
		}

		if err := pgTable.AddRow(cols...); err != nil {
			printCols(cols)
			return err
		}

		success++
	}

	return nil
}
