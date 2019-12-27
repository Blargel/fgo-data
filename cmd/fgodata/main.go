package main

import (
	"database/sql"
	"flag"

	"../../pkg/fgodata"
)

var input string
var dburl string

func init() {
	flag.StringVar(&input, "input", "", "input json file")
	flag.StringVar(&dburl, "dburl", "postgres://localhost/fgo_data?sslmode=disable", "database name to create")
	flag.Parse()
}

func main() {
	if input == "" || dburl == "" {
		flag.PrintDefaults()
		return
	}

	fgo, err := fgodata.ImportData(input)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", dburl)
	if err != nil {
		panic(err)
	}

	err = fgo.ResetSchema(db)
	if err != nil {
		panic(err)
	}

	err = fgo.InsertData(db)
	if err != nil {
		panic(err)
	}
}
