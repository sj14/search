package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	c = make(chan string, 100) // Allocate a channel.
)

func main() {
	// Connect to Database
	db, err := sql.Open("mysql", "root:@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// get first url to crawl
	c <- popToCrawlURL(db)

	for url := range c {
		crawl(db, url)

		// get next url to crawl
		c <- popToCrawlURL(db)
		time.Sleep(1 * time.Second) // should be a more polite value
	}

}

func crawl(db *sql.DB, url string) {

	fmt.Println("Trying to crawl: ", url)

	var s, err = getBody(url)
	if err != nil {
		return
	}

	// find links
	urlsFound := findLinks(s)

	for _, urlFound := range urlsFound {
		// normalize url
		urlFound, err := normalize(url, urlFound)
		if err != nil {
			fmt.Println(err)
			return
		}

		// insert into "to_crawl" table of db
		insertToCrawlURL(db, urlFound)
		fmt.Println("Found new url: ", urlFound)
	}
}

func getBody(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "", err
	}
	return string(body), nil
}

func findLinks(s string) []string {
	var urlsFound []string

	for cnt := strings.Count(s, "href=\""); cnt > 0; cnt-- {
		start := strings.Index(s, "href=\"") + 6
		if start == -1 {
			break
		}
		s = s[start:]
		end := strings.Index(s, "\"")
		if end == -1 {
			break
		}
		urlFound := s[:end]
		urlsFound = append(urlsFound, urlFound)
	}
	return urlsFound
}

func popToCrawlURL(db *sql.DB) string {
	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id, url FROM to_crawl LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var id int
	var url string // we "scan" the result in here

	// Query the first element found
	err = stmtOut.QueryRow().Scan(&id, &url) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	// Prepare statement for deleting data
	stmtDel, err := db.Prepare("DELETE FROM to_crawl WHERE id = ?") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtDel.Close() // Close the statement when we leave main() / the program terminates,

	// Delete the element
	_, err = stmtDel.Exec(id)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	return url
}

func insertToCrawlURL(db *sql.DB, url string) {
	// Connect to Database
	db, err := sql.Open("mysql", "root:@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO to_crawl (url) VALUES(?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database

	_, err = stmtIns.Exec(url) // Insert tuples (i, i^2)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Print(err)
	}

}

func normalize(urlStart, urlFound string) (string, error) {
	// Add http if protocol isn't set
	if len(urlFound) > 1 && urlFound[:2] == "//" {
		urlFound = "http:" + urlFound
	}
	// Set start url in front if it's not set
	if len(urlFound) > 1 && urlFound[:1] == "/" {
		urlFound = urlStart + urlFound
	}
	// only add http(s) links
	if len(urlFound) > 7 && urlFound[0:7] != "http://" {
		if len(urlFound) > 8 && urlFound[0:8] != "https://" {
			return "", errors.New("Protocol should be http(s)")
		}
	}
	return urlFound, nil
}
