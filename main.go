package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type Movie struct {
	Title string `json:"Title"`
	Year int16 `json:"year"`
	Director string `json:"director"`
}

// global Movies array to simulate a database, for now
var Movies []Movie

func homePage(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Welcome to the Homepage!")
	fmt.Println("In: homePage")
}

func returnAllMovies(w http.ResponseWriter, r *http.Request){
	fmt.Println("In: returnAllMovies")
	json.NewEncoder(w).Encode(Movies)
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", homePage)
	router.HandleFunc("/all", returnAllMovies)

	log.Fatal(http.ListenAndServe(":8081", router))
}

func main() {
	fmt.Println("Rest API started...")
	Movies = []Movie{
		{Title: "The Royal Tenenbaums", Year: 2001, Director: "Wes Anderson"},
		{Title: "The Italian Job", Year: 2003, Director: "F. Gary Gray"},
	}
	handleRequests()
}
