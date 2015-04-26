package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	//urls  []string
	wg sync.WaitGroup
	c  = make(chan string, 100) // Allocate a channel.

)

func main() {

	// Database
	db, err := sql.Open("mysql", "root:asd@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT url FROM to_crawl LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var url string // we "scan" the result in here

	// Query the square-number of 13
	err = stmtOut.QueryRow().Scan(&url) // WHERE number = 13
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	//c <- "https://plus.google.com/+LinusTorvalds/posts"
	c <- url

	for i := 0; i < 10; i++ {

		wg.Add(1)

		go crawl(<-c)

	}

	wg.Wait()

}

func crawl(url string) {
	defer wg.Done()

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

		// add url to list of not yet crawled urls
		// has to run in own goroutine, otherwise causes a deadlock
		go func() {
			c <- urlFound
		}()
		fmt.Println("Found new url: ", urlFound)

	}
}

func parseLinks() {

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
