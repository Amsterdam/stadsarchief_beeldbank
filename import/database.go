package main

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/kelseyhightower/envconfig"
	"github.com/lib/pq"
)

type DBConfigSpec struct {
	Debug    bool   `default:"false"`
	Port     int    `default:"5432"`
	User     string `default:"beeldbank"`
	Password string `default:"insecure"`
	Database string `default:"beeldbank"`
	Host     string `default:"database"`
}

var (
	DBConfig 	DBConfigSpec
)

//	ConnectStr create string to connect to database
func connectStr(prefix string) string {
	err := envconfig.Process(prefix, &DBConfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	otherParams := "sslmode=disable connect_timeout=5"
	connectString := fmt.Sprintf(
		"user=%s dbname=%s password='%s' host=%s port=%d %s",
		DBConfig.User,
		DBConfig.Database,
		DBConfig.Password,
		DBConfig.Host,
		DBConfig.Port,
		otherParams,
	)
	fmt.Println("Connecting..:", connectString)
	return connectString
}

//	setup a database connection
func DBConnect(prefix string) *gorm.DB {
	//db, err := sql.Open("postgres", connectStr)
	db, err := gorm.Open("postgres", connectStr(prefix))

	if err != nil {
		panic(err.Error())
	}

	return db
}

//	close a database connection
func DBClose(DB *gorm.DB) {
	err := DB.Close()
	if err != nil {
		panic(err.Error())
	}
}


//Migrate add missing tables to database
func Migrate(DB *gorm.DB) {
	log.Printf("Create db tables..")
	DB.DropTableIfExists(&ImageLocation{}, &BeeldbankImage{}, &ImageFileLocation{})
	DB.AutoMigrate(&ImageLocation{}, &BeeldbankImage{}, &ImageFileLocation{})
}

//	SQLImport import structure
type SQLImport struct {
	txn  *sql.Tx
	stmt *sql.Stmt
}

//	AddRow Add a single row of data to the database
func (i *SQLImport) AddRow(columns ...interface{}) error {
	_, err := i.stmt.Exec(columns...)
	return err
}

//	Commit the import to database
func (i *SQLImport) Commit() error {
	_, err := i.stmt.Exec()
	if err != nil {
		panic(err)
	}

	// Statement might already be closed therefore ignore errors
	_ = i.stmt.Close()

	return i.txn.Commit()
}

//	NewImport setup a new import struct
func newImport(db *sql.DB, schema string, tableName string, columns []string) (*SQLImport, error) {
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
		if field == "" {
			cols[i] = nil
			continue
		}
		cols[i] = field
	}

	return cols, nil
}

//	print columns we try to put in database
func printCols(cols []interface{}) {
	log.Printf("columns:")
	for i, field := range cols {
		log.Printf("%2d %-250s", i, field)
	}
}

//	StreamInTable data from channel into specified database table.
func StreamInTable(tablename string, columns []string, rows <-chan *[]string, DB *gorm.DB, wg *sync.WaitGroup) {
	cdb := DB.CommonDB().(*sql.DB)
	defer cdb.Close()
	defer wg.Done()

	pgTable, err := newImport(cdb, "public", tablename, columns)

	if err != nil {
		panic(err)
	}

	for row := range rows {
		cols, err := normalizeRow(row)

		if err != nil {
			log.Println(err)
			continue
		}

		if err := pgTable.AddRow(cols...); err != nil {
			printCols(cols)
			panic(err)
		}
	}

	log.Println("DONE! inserting records in", tablename)
	err = pgTable.Commit()

	if err != nil {
		panic(err)
	}
}
