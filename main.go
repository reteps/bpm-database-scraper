package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os"
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
	year INT,
	combo TEXT UNIQUE
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
	genre  string
}

func main() {
	// Load environment
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	// Connect to database

	info := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, os.Getenv("USER"), os.Getenv("PASSWORD"), db)
	db, err := sql.Open("postgres", info)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	baseURL := "https://www.beatport.com/tracks/all"
	bpmURL := baseURL + "?bpm-low=%d&bpm-high=%d"
	pageChan := make(chan category)
	songChan := make(chan song)
	bpmStart := 47
	bpmEnd := 258

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
	for i := bpmStart; i < bpmEnd; i++ {
		url := fmt.Sprintf(bpmURL, i, i)
		go getPageCount(netClient, pageChan, url)
	}
	count := 0
	for {
		select {
		case category := <-pageChan:
			for i := 1; i <= category.count; i++ {
				subURL := fmt.Sprintf("%s&page=%d", category.baseURL, i)
				time.Sleep(50 * time.Millisecond)
				go getCategoryPage(netClient, songChan, subURL)
			}

		// case artist := <-artistChan:
		// 	time.Sleep(50 * time.Millisecond)
		// 	go getArtistPage(netClient, songChan, artist.url)

		case song := <-songChan:
			fmt.Println(song.title, "|", song.artist, "|", song.bpm, "|", song.genre, "|", song.year)
			// sqlStatement := `
			// INSERT INTO songs (song, artist, bpm, year, genre)
			// VALUES ($1, $2, $3, $4, $5)`
			// // defer db.Close()
			// year := song.year
			// if year == "—" || year == "-" || year == "" {
			// 	year = "0"
			// }
			// bpm := song.bpm
			// if bpm == "—" || bpm == "-" || bpm == "" {
			// 	bpm = "0"
			// }
			// _, err = db.Exec(sqlStatement, song.title, song.artist, bpm, year, song.title+" "+song.artist)
			// if err != nil && strings.Contains(err.Error(), "invalid") {
			// 	fmt.Println(song.title, song.artist, song.bpm, song.year)
			// 	panic(err)
			// } else {
			count++
			fmt.Println(song.title, count)
			// }
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
	doc, err := getPage(client, url)
	if err != nil {
		if err.Error() == "502" || err.Error() == "500" || strings.Contains(err.Error(), "no such host") {
			fmt.Printf("500/502 Error, retrying %s...\n", url)
			time.Sleep(50 * time.Millisecond)
			getPageCount(client, pageChan, url)
			return
		}
		panic(err)
		// return
	}
	pages := doc.Find("div.pag-numbers > a.pag-number").Last().Text()
	if pages == "" {
		pages = "1"
	}
	// if !success {
	// 	panic(fmt.Errorf("Could not find the number of pages for url %s", url))
	// }
	// fmt.Println(url, pages)
	if err != nil {
		panic(err)
	}
	val, _ := strconv.Atoi(pages)
	pageChan <- category{url, val}

}

// Take a category page with artists and send all artists to artistChan
func getCategoryPage(client http.Client, songChan chan song, url string) {
	// fmt.Println("url=", url)
	doc, err := getPage(client, url)
	if err != nil {
		if err.Error() == "502" || err.Error() == "500" || strings.Contains(err.Error(), "no such host") {
			fmt.Printf("500/502 Error, retrying %s...\n", url)
			time.Sleep(50 * time.Millisecond)
			getCategoryPage(client, songChan, url)
			return
		}
		panic(err)
		// return
	}
	doc.Find("li.bucket-item").Each(func(i int, item *goquery.Selection) {
		name, success := item.Attr("data-ec-name")
		genre, success := item.Attr("data-ec-d3")
		artists, success := item.Attr("data-ec-d1")
		year := item.Find("p.buk-track-released").Text()
		if !success {
			panic(fmt.Errorf("Could not find href for %s", url))
		}
		section := strings.Split(url, "&bpm-high=")[1]
		bpm := strings.Split(section, "&")[0]
		songChan <- song{artists, name, bpm, year, genre}

	})
}
