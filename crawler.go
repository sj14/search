package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	//urls  []string
	//wg sync.WaitGroup
	c = make(chan string, 100) // Allocate a channel.

)

func main() {
	c <- popToCrawlURL()

	for url := range c {
		crawl(url)
		c <- popToCrawlURL()
		time.Sleep(1 * time.Second) // should be a more polite value
	}

}

func crawl(url string) {

	fmt.Println("Trying to crawl: ", url)

	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	var s = string(body)

	// find links
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

		// normalize url
		urlFound, err = normalize(url, urlFound)
		if err != nil {
			fmt.Println(err)
			return
		}

		insertToCrawlURL(urlFound)
		fmt.Println("Found new url: ", urlFound)

	}
}

func popToCrawlURL() string {
	// Connect to Database
	db, err := sql.Open("mysql", "root:asd@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

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

func insertToCrawlURL(url string) {
	// Connect to Database
	db, err := sql.Open("mysql", "root:asd@/search")
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
		panic(err.Error()) // proper error handling instead of panic in your app
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
