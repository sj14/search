package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"text/template"
)

func handlerIndex(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	body, _ := ioutil.ReadFile("index.html")
	t.Execute(w, body)
	fmt.Fprintf(w, "Hi Index, I love %s!", r.URL.Path[1:])
}

func handlerQuery(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	body, _ := ioutil.ReadFile("index.html")
	t.Execute(w, body)
	fmt.Fprintf(w, "Hi Query, I love %s!", r.URL.Path[1:])
}

func main() {
	http.HandleFunc("/", handlerIndex)
	http.HandleFunc("/query", handlerQuery)
	http.ListenAndServe(":8080", nil)
}
