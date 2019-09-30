package main

import (
	"github.com/PuerkitoBio/goquery"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
)

/*
/Applications/Postgres.app/Contents/Versions/11/bin/psql -p5432 "407869"
CREATE DATABASE calhounio_demo;
\c database

CREATE TABLE songs (
	song TEXT,
	artist TEXT,
	bpm INT,
	year INT
)
*/
const (
	host     = "localhost"
	port     = 5432
	user     = "407869"
	password = "Retep2170!"
	db       = "song_table"
	table    = "songs"
)
type Category struct {
	base string
	count int
}
func main() {

	fmt.Println("Hello world!")
	info := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, db)
	db, err := sql.Open("postgres", info)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection successful")

	// sqlStatement := `
	// INSERT INTO $1 (song, artist, bpm, year)
	// VALUES ($2, $3, $4, $5)`
	// _, err = db.Exec(sqlStatement)
	defer db.Close()

	baseURL := "https://bpmdatabase.com/music/%s"
	rawCategories := append(strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", ""), "0-9")
	var 
	for _, letter := range rawCategories {
		url := fmt.Sprintf(baseURL, letter)
		fmt.Println(url)
	}

}
func getPage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // When we read from body, make sure to close it
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status code is %d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
func insertSongs(url, db *sql.DB) {

}
