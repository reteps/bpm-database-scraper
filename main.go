package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type category struct {
	baseURL string
	count   int
}
type artist struct {
	url  string
	name string
}
type song struct {
	artist string
	title  string
	bpm    string
	year   string
}

func main() {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	// Connect to database

	// info := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, os.Getenv("USER"), os.Getenv("PASSWORD"), db)
	// db, err := sql.Open("postgres", info)
	// if err != nil {
	// 	panic(err)
	// }
	// err = db.Ping()
	// if err != nil {
	// 	panic(err)
	// }

	// sqlStatement := `
	// INSERT INTO $1 (song, artist, bpm, year)
	// VALUES ($2, $3, $4, $5)`
	// _, err = db.Exec(sqlStatement)
	// defer db.Close()

	baseURL := "https://bpmdatabase.com/music/%s"
	rawCategories := append(strings.Split("ABCDEFGHIJKLMNOPQRSTUVWXYZ", ""), "0-9")
	pageChan := make(chan category)
	artistChan := make(chan artist)
	songChan := make(chan song)
	dialer := (&net.Dialer{
		Timeout: 10 * time.Second,
	}).DialContext
	netClient := http.Client{
		Transport: &http.Transport{
			DialContext:         dialer,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	// make a request to each letter URL
	for _, letter := range rawCategories {
		url := fmt.Sprintf(baseURL, letter)
		go getPageCount(netClient, pageChan, url)
	}
	count := 0
	for {
		select {
		case category := <-pageChan:
			for i := 1; i <= category.count; i++ {
				subURL := fmt.Sprintf("%s?page=%d", category.baseURL, i)
				go getCategoryPage(netClient, artistChan, subURL)
			}

		case artist := <-artistChan:
			go getArtistPage(netClient, songChan, artist.url)

		case song := <-songChan:
			count++
			fmt.Println(song.title, count)
		}
	}
}

// Generic method to take a URL and return a goquery Doc
func getPage(client http.Client, url string) (*goquery.Document, error) {
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close() // When we read from body, make sure to close it
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%d", res.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Take a category page and send the category and page count to pageChan
func getPageCount(client http.Client, pageChan chan category, url string) {
	time.Sleep(50 * time.Millisecond)
	doc, err := getPage(client, url)
	if err != nil {
		if err.Error() == "502" || err.Error() == "500" {
			fmt.Printf("500/502 Error, retrying %s...\n", url)
			getPageCount(client, pageChan, url)
			return
		}
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
	pageChan <- category{url, pageNum}

}

// Take a category page with artists and send all artists to artistChan
func getCategoryPage(client http.Client, artistChan chan artist, url string) {
	time.Sleep(50 * time.Millisecond)
	doc, err := getPage(client, url)
	if err != nil {
		if err.Error() == "502" || err.Error() == "500" {
			fmt.Printf("500/502 Error, retrying %s...\n", url)
			getCategoryPage(client, artistChan, url)
			return
		}
		log.Println(err)
		return
	}
	doc.Find("div.list-group > a").Each(func(i int, link *goquery.Selection) {
		artistURL, success := link.Attr("href")
		name := link.Text()
		if !success {
			panic(fmt.Errorf("Could not find href for %s", url))
		}
		artistChan <- artist{"https://bpmdatabase.com" + artistURL, name}
	})
}

// Take an artists page and send all of their songs to songChan
func getArtistPage(client http.Client, songChan chan song, url string) {
	time.Sleep(50 * time.Millisecond)
	doc, err := getPage(client, url)
	if err != nil {
		if err.Error() == "502" || err.Error() == "500" {
			fmt.Printf("500/502 Error, retrying %s...\n", url)
			getArtistPage(client, songChan, url)
			return
		}
		log.Println(err)
		return
	}
	doc.Find("tbody > tr").Each(func(i int, row *goquery.Selection) {
		artist := row.Find(".artist").Text()
		title := row.Find(".title").Text()
		bpm := row.Find(".bpm").Text()
		year := row.Find(".year").Text()
		songChan <- song{artist, title, bpm, year}
	})
}
