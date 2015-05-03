package main

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	c = make(chan url.URL, 100) // Allocate a channel.
)

func main() {
	// Connect to Database
	db, err := sql.Open("mysql", "root:@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// get first urlarg to crawl
	//c <- popToCrawlURL(db)

	startURL, _ := url.Parse("https://www.udacity.com/cs101x/index.html")

	c <- *startURL

	for urlarg := range c {
		crawl(db, urlarg)

		// get next url to crawl
		c <- popToCrawlURL(db)
		time.Sleep(1 * time.Second) // should be a more polite value
	}
}

func crawl(db *sql.DB, urlarg url.URL) {

	log.Println("Trying to crawl: ", urlarg)

	var s, err = getBody(urlarg)
	if err != nil {
		return
	}

	insertBodyToTableURL(db, urlarg, s)

	// find links
	urlsFound := findLinks(s)

	for _, urlFound := range urlsFound {
		// normalize urlarg

		urlFound, err := normalize(urlarg, urlFound)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("Found new url in body: ", urlFound)
		// insert into "to_crawl" table of db
		insertToCrawlURL(db, urlFound)
	}
}

func getBody(urlarg url.URL) (string, error) {
	resp, err := http.Get(urlarg.String())
	if err != nil {
		log.Println(err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return string(body), nil
}

func findLinks(s string) []url.URL {
	var urlsFound []url.URL

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
		urlParsedFound, err := url.Parse(urlFound)
		if err != nil {
			log.Println(err)
		}

		urlsFound = append(urlsFound, *urlParsedFound)
	}
	return urlsFound
}

func popToCrawlURL(db *sql.DB) url.URL {

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id, url FROM to_crawl LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var id int
	var urlarg string // we "scan" the result in here

	// Query the first element found
	err = stmtOut.QueryRow().Scan(&id, &urlarg) // WHERE number = 13
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println("No more URLs to crawl. Exiting.")
		os.Exit(0)
	}

	parsedURL, err := url.Parse(urlarg)
	if err != nil {
		log.Fatal(err)
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
	return *parsedURL
}

func insertToCrawlURL(db *sql.DB, urlarg url.URL) {

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT url FROM urls WHERE url = ? LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	//var dupURL string

	// Query the first element found
	err = stmtOut.QueryRow(urlarg.String()).Scan() // WHERE number = 13
	if err != nil {
		log.Printf("prevented adding already crawled url: %v", urlarg.String())
		return
	}

	// if dupURL != "" {
	// 	fmt.Println("prevented adding already crawled url")
	// 	return
	// }

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO to_crawl (url) VALUES(?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database
	_, err = stmtIns.Exec(urlarg.String()) // Insert tuples (i, i^2)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println(err)
	}

}

func insertBodyToTableURL(db *sql.DB, urlarg url.URL, body string) {
	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO urls (url, text) VALUES(?, ?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database

	_, err = stmtIns.Exec(urlarg.String(), body) // Insert tuples (i, i^2)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println(err)
	}

}

func normalize(urlargStart, urlFound url.URL) (url.URL, error) {
	// // Add http if protocol isn't set
	// if len(urlFound) > 1 && urlFound[:2] == "//" {
	// 	urlFound = "http:" + urlFound
	// }
	// // Set start urlarg in front if it's not set
	// if len(urlFound) > 1 && urlFound[:1] == "/" {
	// 	urlFound = urlargStart + urlFound
	// }
	// // only add http(s) links
	// if len(urlFound) > 7 && urlFound[0:7] != "http://" {
	// 	if len(urlFound) > 8 && urlFound[0:8] != "https://" {
	// 		return "", errors.New("Protocol should be http(s)")
	// 	}
	// }

	// Add protocol if blank
	if urlFound.Scheme == "" {
		urlFound.Scheme = urlargStart.Scheme
	}

	// Add host if blank
	if urlFound.Host == "" {
		urlFound.Host = urlargStart.Host
	}

	// only add http(s) links
	if urlFound.Scheme != "http" {
		if urlFound.Scheme != "https" {
			return urlFound, errors.New("Protocol should be http(s)")
		}
	}

	return urlFound, nil
}
