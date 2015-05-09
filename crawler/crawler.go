package main

import (
	"bufio"
	"database/sql"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	c           = make(chan url.URL, 25) // Allocate channel(s).
	lastCrawled = make(map[string]time.Time)
	mutex       = &sync.Mutex{}
)

func main() {
	// Connect to Database
	db, err := sql.Open("mysql", "root:@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	// get first urlArg to crawl
	//c <- popToCrawlURL(db)

	//startURL, _ := url.Parse("https://www.udacity.com/cs101x/index.html")

	startURL, _ := url.Parse("http://golang.com/")
	c <- *startURL

	for i := 0; i < 24; i++ {
		c <- getCrawlURL(db)
	}

	for urlArg := range c {
		time.Sleep(10 * time.Millisecond)
		go handleCrawl(db, urlArg)
	}
}

func handleCrawl(db *sql.DB, urlArg url.URL) {
	checkDelay(urlArg)

	mutex.Lock()
	lastCrawled[urlArg.Host] = time.Now()
	mutex.Unlock()

	crawl(db, urlArg)
	// get next url to crawl

	c <- getCrawlURL(db)
}

func checkDelay(urlArg url.URL) {
	mutex.Lock()
	lastTime := lastCrawled[urlArg.Host]
	mutex.Unlock()
	timeSince := time.Since(lastTime)

	if timeSince.Seconds() < 1 {
		waitingTime := 1*time.Second - timeSince
		//log.Printf("Waiting %v before crawling %v again", waitingTime, urlArg.Host)
		time.Sleep(waitingTime) // should be a more polite value

		checkDelay(urlArg)
	}

}

func crawl(db *sql.DB, urlArg url.URL) {
	log.Println("Trying to crawl: ", urlArg)
	var s, err = getBody(urlArg)
	if err != nil {
		log.Println(err)
		return
	}

	//insertURLToDB(db, urlArg)
	inserKeywordsToDB(db, urlArg, s)

	// find links
	urlsFound := findLinks(s)

	for _, urlFound := range urlsFound {
		// normalize url
		urlFound, err := normalize(urlArg, urlFound)
		if err != nil {
			log.Println(err)
			return
		}
		//log.Println("Found new url in body: ", urlFound)
		// insert into "to_crawl" table of db
		insertToCrawlURL(db, urlFound)
	}
}

func getBody(urlArg url.URL) (string, error) {
	respHead, err := http.Head(urlArg.String())
	if err != nil {
		return "", err
	}
	//log.Println(respHead.Header.Get("Content-Type"))
	contentType := respHead.Header.Get("Content-Type")

	if !strings.Contains(contentType, "text/html") {
		return "", errors.New("Not text/html content-type")
	}

	resp, err := http.Get(urlArg.String())
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
			continue
		}
		s = s[start:]
		end := strings.Index(s, "\"")
		if end == -1 {
			continue
		}
		urlFound := s[:end]
		urlParsedFound, err := url.Parse(urlFound)
		if err != nil {
			log.Println(err)
			continue
		}
		// Filter for interesting links (mostly html)
		if strings.HasSuffix(urlParsedFound.Path, "/") || strings.HasSuffix(urlParsedFound.Path, ".html") {
			urlsFound = appendURLIfMissing(urlsFound, *urlParsedFound)
		}
	}
	return urlsFound
}

// Read first url from DB, save it into variable and remove from DB
func getCrawlURL(db *sql.DB) url.URL {
	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
	}

	// Prepare statement for first url
	stmtOut, err := tx.Prepare("SELECT url FROM crawl  WHERE timestamp IS NULL LIMIT 1")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	var urlArg string // we "scan" the result in here

	// Query the first element found
	err = stmtOut.QueryRow().Scan(&urlArg) // WHERE number = 13
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println("No more URLs to crawl. Exiting.")
		//os.Exit(0)
	}

	parsedURL, err := url.Parse(urlArg)
	if err != nil {
		log.Fatal(err)
	}

	// Prepare statement for inserting data
	stmtIns, err := tx.Prepare("UPDATE crawl SET timestamp = ? WHERE url = ?") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database
	_, err = stmtIns.Exec(time.Now(), urlArg) // Insert tuples (i, i^2)
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
	}
	return *parsedURL
}

// check if url has already been crawled (if it is in table 'urls'!)
// and if not, add to to_crawl table. The db will check if it is already in to_crawl
func insertToCrawlURL(db *sql.DB, urlArg url.URL) {

	/*
		// Prepare statement for reading data
		stmtOut, err := db.Prepare("SELECT url FROM crawl WHERE url = ? LIMIT 1")
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		defer stmtOut.Close()

		// Query the first element found
		err = stmtOut.QueryRow(urlArg.String()).Scan() // WHERE number = 13
		// no error means the url has been found in the db
		if err == nil {
			log.Printf("prevented adding already crawled url: %v", urlArg.String())
			return
		}
	*/

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO crawl (url) VALUES(?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Insert square numbers for 0-24 in the database
	_, err = stmtIns.Exec(urlArg.String()) // Insert tuples (i, i^2)
	if err != nil {
		// Propably duplicate entry
		//log.Println(err)
	}
}

// insert text/body from the website to db table 'urls'
/*
func insertURLToDB(db *sql.DB, urlArg url.URL) {
	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO urls (url) VALUES(?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	_, err = stmtIns.Exec(urlArg.String()) // Insert tuples (i, i^2)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		log.Println(err)
	}
}
*/

func inserKeywordsToDB(db *sql.DB, urlArg url.URL, body string) {
	// save already inserted keywords to reduce db load
	uniqueWords := make(map[string]bool)

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO keyword_url (fk_keyword, fk_url) VALUES(?, ?)") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	scanner := bufio.NewScanner(strings.NewReader(body))
	// Set the split function for the scanning operation.
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {

		// Ignores all words < 3 chars
		// Ignores a lot html keywords and of course valid ones...
		if len(scanner.Text()) >= 3 && !strings.HasPrefix(scanner.Text(), "<") {

			ok := uniqueWords[scanner.Text()]

			if !ok {
				uniqueWords[scanner.Text()] = true

				_, err = stmtIns.Exec(scanner.Text(), urlArg.String()) // Insert tuples (i, i^2)
				if err != nil {
					// Propably duplicate entry
					//panic(err.Error()) // proper error handling instead of panic in your app
					//log.Println(err)
				}
			}
		}
	}
}

func normalize(urlArgStart, urlFound url.URL) (url.URL, error) {
	// Add protocol if blank
	if urlFound.Scheme == "" {
		urlFound.Scheme = urlArgStart.Scheme
	}

	// Add host if blank
	if urlFound.Host == "" {
		urlFound.Host = urlArgStart.Host
	}

	// Remove fragements/anchors -> '#'
	if urlFound.Fragment != "" {
		urlFound.Fragment = ""
	}

	// Remove queries -> '?'
	if urlFound.RawQuery != "" {
		urlFound.RawQuery = ""
	}

	// only add http(s) links
	if urlFound.Scheme != "http" {
		if urlFound.Scheme != "https" {
			return urlFound, errors.New("Protocol should be http(s)")
		}
	}
	return urlFound, nil
}

// From http://stackoverflow.com/a/9561388
func appendURLIfMissing(slice []url.URL, u url.URL) []url.URL {
	for _, ele := range slice {
		if ele == u {
			return slice
		}
	}
	return append(slice, u)
}

func appendStringIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}
