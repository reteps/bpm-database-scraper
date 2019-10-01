package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

/*
/Applications/Postgres.app/Contents/Versions/11/bin/psql -p5432 "USER"
CREATE DATABASE song_table;
\c database

CREATE TABLE songs (
	song TEXT,
	artist TEXT,
	bpm INT,
	year INT
);
*/
const (
	host  = "localhost"
	port  = 5432
	db    = "song_table"
	table = "songs"
)

type Category struct {
	baseURL string
	count   int
}
type Artist struct {
	url  string
	name string
}
type Song struct {
	artist string
	title  string
	bpm    string
	year   string
}

func main() {
	err := godotenv.Load()
	info := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, os.Getenv("USER"), os.Getenv("PASSWORD"), db)
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

	pageChan := make(chan Category)
	artistChan := make(chan Artist)
	songChan := make(chan Song)
	for _, letter := range rawCategories {
		url := fmt.Sprintf(baseURL, letter)
		go getPageCount(pageChan, url)
	}
	for {
		select {
		case category := <-pageChan:
			for i := 1; i <= category.count; i++ {
				subURL := fmt.Sprintf("%s?page=%d", category.baseURL, i)
				go getCategoryPage(artistChan, subURL)

			}
		case artist := <-artistChan:
			go getArtistPage(songChan, artist.url)
		case song := <-songChan:
			fmt.Println(song.title)
		}

	}
}

func getPage(url string) (*goquery.Document, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // When we read from body, make sure to close it
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Status code is %d for %s", res.StatusCode, url)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
func getArtistPage(songChan chan Song, url string) {
	doc, err := getPage(url)
	if err != nil {
		log.Println(err)
		return
	}
	doc.Find("tbody > tr").Each(func(i int, row *goquery.Selection) {
		artist := row.Find(".artist").Text()
		title := row.Find(".title").Text()
		bpm := row.Find(".bpm").Text()
		year := row.Find(".year").Text()
		songChan <- Song{artist, title, bpm, year}
	})
}
func getPageCount(pageChan chan Category, url string) {
	doc, err := getPage(url)
	if err != nil {
		log.Println(err)
		return
	}
	pages, success := doc.Find("ul.pagination > li.last > a").Attr("href")
	if !success {
		panic(fmt.Errorf("Could not find the number of pages for url %s", url))
	}
	pageNum, err := strconv.Atoi(strings.Split(pages, "=")[1])
	if err != nil {
		panic(err)
	}
	pageChan <- Category{url, pageNum}

}
func getCategoryPage(artistChan chan Artist, url string) {
	doc, err := getPage(url)
	if err != nil {
		log.Println(err)
		return
	}
	doc.Find("div.list-group > a").Each(func(i int, link *goquery.Selection) {
		artistURL, success := link.Attr("href")
		name := link.Text()
		if !success {
			panic(fmt.Errorf("Could not find href for %s", url))
		}
		artistChan <- Artist{"https://bpmdatabase.com" + artistURL, name}
	})
}
