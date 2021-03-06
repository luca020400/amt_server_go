package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/anaskhan96/soup"
	"github.com/gorilla/mux"
)

var (
	port = "5555"

	urlLine = "https://www.amt.genova.it/amt/servizi/orari_tel.php"
	urlStop = "https://www.amt.genova.it/amt/simon.php?CodiceFermata="
)

type StopData struct {
	Line string `json:"line"`
	Dest string `json:"dest"`
	Time string `json:"time"`
	ETA  string `json:"eta"`
}

type Stop struct {
	Name  string     `json:"name"`
	Stops []StopData `json:"stops"`
}

type LineData struct {
	Direction string   `json:"direction"`
	Times     []string `json:"times"`
}

type Line struct {
	Lines []LineData `json:"lines"`
}

func downloadStop(code string) []byte {
	resp, err := http.Get(urlStop + code)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	return body
}

func parseStop(html []byte) []byte {
	doc := soup.HTMLParse(string(html))

	names := doc.FindAll("font")
	if len(names) != 3 {
		// Non existing stop, craft an empty json
		js, err := json.Marshal(Stop{})
		if err != nil {
			// handle error
		}
		return js
	}

	var stopData []StopData
	for _, tr := range doc.FindAll("tr") {
		tds := tr.FindAll("td")
		if len(tds) == 4 {
			stopData = append(stopData, StopData{tds[0].Text(), tds[1].Text(), tds[2].Text(), tds[3].Text()})
		}
	}

	name := names[1].Text()

	js, err := json.Marshal(Stop{name, stopData})
	if err != nil {
		// handle error
	}

	return js
}

func downloadLine(line string) []byte {
	today := time.Now()

	formData := url.Values{
		"giorno":   {strconv.Itoa(today.Day())},
		"mese":     {strconv.Itoa(int(today.Month()) - 1)},
		"anno":     {strconv.Itoa(today.Year())},
		"linea":    {line},
		"cmdOrari": {"Mostra Orari"},
	}

	resp, err := http.PostForm(urlLine, formData)
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}

	return body
}

func parseLine(html []byte) []byte {
	doc := soup.HTMLParse(string(html))

	var lineData []LineData

	bs := doc.FindAll("b")
	directions := bs[:0]
	for _, b := range bs {
		if !strings.HasPrefix(b.Text(), "LINEA") {
			directions = append(directions, b)
		}
	}

	tables := doc.FindAll("table")
	timeTables := tables[:0]
	for _, table := range tables {
		if len(table.FindAll("td")) != 0 {
			timeTables = append(timeTables, table)
		}
	}

	for index, table := range timeTables {
		var times []string
		for _, td := range table.FindAll("td") {
			times = append(times, td.Text())
		}
		lineData = append(lineData, LineData{directions[index].Text(), times})
	}

	js, err := json.Marshal(Line{lineData})
	if err != nil {
		// handle error
	}

	return js

}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		next.ServeHTTP(w, r)
	})
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}

func StopHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page := downloadStop(mux.Vars(r)["stop"])
	js := parseStop(page)
	w.Write(js)
}

func LineHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	page := downloadLine(mux.Vars(r)["line"])
	js := parseLine(page)
	w.Write(js)
}

func main() {
	router := mux.NewRouter()
	api := router.PathPrefix("/api").Subrouter()
	api.NotFoundHandler = http.HandlerFunc(NotFoundHandler)
	api.Use(Middleware)

	api1 := api.PathPrefix("/v1").Subrouter()
	api1.HandleFunc("/stop/{stop:\\d{4}}", StopHandler)
	api1.HandleFunc("/line/{line:[A-Z0-9]{1,3}}", LineHandler)
	api1.NotFoundHandler = http.HandlerFunc(NotFoundHandler)

	log.Println("Listening on port", port)
	http.ListenAndServe(":"+port, router)
}
