package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	wd "github.com/settermjd/weatherdata"

	"github.com/gorilla/mux"
	_ "github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

type DisclaimerPageData struct {
	PageTitle string
}

type DefaultPageData struct {
	PageTitle   string
	WeatherData []wd.WeatherData
}

func BuildTemplate(templateName string) *template.Template {
	log.Print("Content template is: " + templateName)
	tmpl := template.Must(
		template.ParseFiles(
			filepath.Join("templates", "routes", templateName+".html"),
			filepath.Join("templates", "footer.html"),
			filepath.Join("templates", "layout.html"),
		))

	return tmpl
}

type HandlerContext struct {
	wds *wd.WeatherDataService
}

func NewHandlerContext(wds *wd.WeatherDataService) *HandlerContext {
	if wds == nil {
		panic("nil Weather Data object")
	}
	return &HandlerContext{wds}
}

func (ctx *HandlerContext) DefaultRouteHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := BuildTemplate("default")
	err := tmpl.ExecuteTemplate(w, "layout", DefaultPageData{
		PageTitle:   "DIY Weather Station",
		WeatherData: ctx.wds.GetWeatherData(wd.WeatherDataSearchParams{}),
	})
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())

		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)

		return
	}
}

// This function handles static routes
// That is, routes that have no dynamic properties and only return a rendered, static HTML template
func (ctx *HandlerContext) HandleStaticRoute(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	path := params["path"]

	tmpl := BuildTemplate(path)
	err := tmpl.ExecuteTemplate(w, "layout", DisclaimerPageData{PageTitle: "Disclaimer"})
	if err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(404), 404)

		return
	}
}

func main() {
	// Initialise a connection to the application's database
	const DatabaseFile = "./data/database/weather_station_test_data.sqlite"
	db, err := sql.Open("sqlite3", DatabaseFile)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	ctx := NewHandlerContext(wd.NewWeatherDataService(db))

	r := mux.NewRouter()

	// Set up the routing table
	r.HandleFunc("/", ctx.DefaultRouteHandler).
		Methods("GET")

	// Serve static pages
	r.HandleFunc(
		"/{path:about|disclaimer|cookie-policy|datenschutzerklaerung|disclaimer|impressum|privacy-policy}",
		ctx.HandleStaticRoute,
	).Methods("GET")

	// Serve static assets
	fs := http.FileServer(http.Dir("./assets"))
	r.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Boot the application
	err = http.ListenAndServe(":8001", r)
	if err != nil {
		log.Println(err.Error())
		return
	}
}
