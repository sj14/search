package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	body, _ := ioutil.ReadFile("index.html")
	t.Execute(w, body)
	fmt.Fprintf(w, "Hi Index, I love %s!", r.URL.Path[1:])
}

func handlerQuery(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:@/search")
	if err != nil {
		panic(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	t, _ := template.ParseFiles("index.html")
	body, _ := ioutil.ReadFile("index.html")
	t.Execute(w, body)
	query := r.URL.Query().Get("search_input")

	keywords := strings.Split(query, " ")

	// Prepare statement for reading data

	prepareSQL := "SELECT fk_url FROM keyword_url as url1 WHERE fk_keyword = ?"
	keywordsInterface := make([]interface{}, len(keywords))
	keywordsInterface[0] = keywords[0]

	if len(keywords) > 1 {
		for val, keyword := range keywords {
			keywordsInterface[val] = keyword
			if val == 0 {
				continue
			} else {
				prepareSQL += " AND fk_url = (select fk_url from keyword_url where fk_keyword = ? and fk_url = url1.fk_url)"
			}
		}
	}

	stmtOut, err := db.Prepare(prepareSQL)
	if err != nil {
		//panic(err.Error()) // proper error handling instead of panic in your app
		fmt.Fprintf(w, err.Error())
		return
	}
	defer stmtOut.Close()

	var url string // we "scan" the result in here

	// Query the square-number of 13
	//err = stmtOut.QueryRow(keywords[0]).Scan(&url) // WHERE number = 13

	rows, err := stmtOut.Query(keywordsInterface...) // WHERE number = 13
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&url)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Fprintf(w, "Results for %v <br> %v", r.URL.Query().Get("search_input"), url)
		fmt.Fprintf(w, "%v <br>", url)

	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	//fmt.Fprintf(w, "Results for %v <br> %v", r.URL.Query().Get("search_input"), url)
}

func main() {
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/query", handlerQuery)
	http.ListenAndServe(":8080", nil)
}
