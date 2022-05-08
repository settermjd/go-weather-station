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

func defaultRoute(w http.ResponseWriter, r *http.Request) {
	db, _ := sql.Open("sqlite3", "./data/database/weather_station_test_data.sqlite")
	rows, _ := db.Query(`SELECT humidity, temperature, timestamp FROM weather_data`)
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

	err := tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())

		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)

		return
	}
}

func main() {
	fs := http.FileServer(http.Dir("assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/", defaultRoute)
	http.ListenAndServe(":8000", nil)
}
