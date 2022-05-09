package main

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

type DefaultPageData struct {
	PageTitle   string
	WeatherData []WeatherData
}

type WeatherData struct {
	Timestamp   time.Time
	Humidity    float32
	Temperature float32
}

type HandlerContext struct {
	db *sql.DB
}

func NewHandlerContext(db *sql.DB) *HandlerContext {
	if db == nil {
		panic("nil database connection")
	}
	return &HandlerContext{db}
}

func (ctx *HandlerContext) DefaultRouteHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := ctx.db.Query(`SELECT humidity, temperature, timestamp FROM weather_data`)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
		return
	}
	defer rows.Close()

	var weatherData []WeatherData
	for rows.Next() {
		var u WeatherData
		err := rows.Scan(&u.Humidity, &u.Temperature, &u.Timestamp)
		if err != nil {
			fmt.Fprintf(w, "Unable to add record: %s\n", err)
		}
		weatherData = append(weatherData, u)
	}

	routeTemplate := filepath.Join("templates", "routes", "default.html")
	footerTemplate := filepath.Join("templates", "footer.html")
	layoutTemplate := filepath.Join("templates", "layout.html")

	tmpl := template.Must(template.ParseFiles(routeTemplate, footerTemplate, layoutTemplate))
	data := DefaultPageData{
		PageTitle:   "DIY Weather Station",
		WeatherData: weatherData,
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())

		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)

		return
	}
}

func main() {
	db, err := sql.Open("sqlite3", "./data/database/weather_station_test_data.sqlite")
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	ctx := NewHandlerContext(db)

	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/", ctx.DefaultRouteHandler)
	err = http.ListenAndServe(":8001", nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
