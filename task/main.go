package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Написать репозиторий к базе данных test на mysql, используя библиотеку sqlx.CREATE TABLE cities (
// 	id INTEGER NOT NULL PRIMARY KEY,
// 	name VARCHAR(30) NOT NULL,
// 	state VARCHAR(30) NOT NULL
// 	);
// 	Реализовать следующие методы:
// 	Create
// 	Delete
// 	Update
// 	List

func (app *application) Create(w http.ResponseWriter, r *http.Request) {
	var city City
	err := json.NewDecoder(r.Body).Decode(&city)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	err = app.CityModel.Insert(&city)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "created city")
}

func (app *application) Update(w http.ResponseWriter, r *http.Request) {
	var city City
	err := json.NewDecoder(r.Body).Decode(&city)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}
	err = app.CityModel.Update(&city)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	fmt.Fprintf(w, "updated city")
}

func (app *application) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	id64, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = app.CityModel.Delete(int64(id64))
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func (app *application) List(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name  string
		Email string
		Filters
	}

	v := NewValidator()

	qs := r.URL.Query()

	input.Filters.Page = readInt(qs, "page", 1, v)
	input.Filters.PageSize = readInt(qs, "page_size", 20, v)

	input.Filters.Sort = readString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "name", "state"}

	if ValidateFilters(v, input.Filters); !v.Valid() {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	users, metadata, err := app.CityModel.GetAll(input.Filters)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]interface{}{"metadata": metadata, "data": users})
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

}

type application struct {
	CityModel *CityModel
}

func main() {

	var dsn string
	flag.StringVar(&dsn, "db-dsn", "postgres://postgres:password@db:5432/kata_test?sslmode=disable", "PostgreSQL DSN")

	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	app := &application{
		CityModel: NewCityModel(db),
	}

	r := chi.NewRouter()
	r.Post("/cities/create", app.Create)
	r.Put("/cities/update", app.Update)
	r.Delete("/cities/delete/{id}", app.Delete)
	r.Get("/cities/list", app.List)

	server := http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	fmt.Println("Starting server on port 8080")

	log.Fatal(server.ListenAndServe())

}
