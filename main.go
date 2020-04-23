package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Movie struct {
	Title string `json:"title"`
	Year string `json:"year"`
	Director string `json:"director"`
	Actors string `json:"actors"`
	Runtime string `json:"runtime"`
	PosterURL string `json:"poster"`
	Plot string `json:"plot"`
	IMDBRating string `json:"imdbRating"`
	Notes string `json:"notes"`
}

// SSL certificate path
var CertFilePath string
// SSL private key path
var KeyFilePath string
// Telegram token
var TelegramToken string
// JotForm Form ID that holds the data
var JotFormFormID string
// JotForm API Key to access the form data
var JotFormAPIKey string
// OMDB API Key to get movie metadata
var OMDBAPIKey string

func homePage(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Welcome to the Homepage!")
	fmt.Println("In: homePage")
}

func returnAllMovies(w http.ResponseWriter, r *http.Request){
	fmt.Println("In: returnAllMovies")
	allMoviesFromJotForm := getAllMovies()

	allMoviesStructured := make([]Movie, 0, len(allMoviesFromJotForm))
	for _, movie := range allMoviesFromJotForm {
		movieStruct := createMovieStructFromJotFormResponse(movie)
		movieStruct = fillMovieStructMetadata(movieStruct)
		allMoviesStructured = append(allMoviesStructured, *movieStruct)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allMoviesStructured)
}

// Create a struct that mimics the webhook request body
// https://core.telegram.org/bots/api#update
type telegramUpdateObj struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
		User struct {
			ID int64 `json:"id"`
			Name string `json:"first_name"`
		} `json:"from"`
	} `json:"message"`
	UpdateID int64 `json:"update_id"`
}

// Create the struct to mimic the send message request format
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

func telegramWebhook(res http.ResponseWriter, req *http.Request) {
	// Decode the JSON body
	body := &telegramUpdateObj{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("Could not decode JSON", err)
		return
	}

	// Pretty print object for now
	s, _ := json.MarshalIndent(body, "", "  ")
	fmt.Println(string(s))

	// Check if the message contains "hello"
	// if not, return without doing anything
	if !strings.Contains(strings.ToLower(body.Message.Text), "hello") {
		return
	}

	randomMovie := getRandomMovie()

	// Send the selected movie to the client
	if err := sendMovieToChat(body.Message.Chat.ID, *randomMovie); err != nil {
		fmt.Println("Error in sending reply:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("Reply sent!")
}

func getAllMovies() []interface{} {
	// Get movies from JotForm
	res, err := http.Get("https://api.jotform.com/form/" + JotFormFormID + "/submissions?apiKey=" + JotFormAPIKey)
	if err != nil {
		log.Fatal(err)
	}

	responseBody, err := ioutil.ReadAll(res.Body)
	var data map[string]interface{}
	err = json.Unmarshal(responseBody, &data)
	if err != nil {
		log.Fatal(err)
	}

	// JotForm API structure
	content := data["content"].([]interface{})

	// should not be so big-sized
	return content
}

func createMovieStructFromJotFormResponse(jotformResponse interface{}) *Movie {
	// sorry future myself
	// JotForm API & json handling in go forced me to do it :(
	jotformFields := jotformResponse.(map[string]interface{})["answers"].(map[string]interface{})

	// Our Movie struct to be returned
	var movie = new(Movie)

	// JotForm API returns every column,
	// need to traverse them all
	for _, answerObj := range jotformFields {
		answerMap := answerObj.(map[string]interface{})

		if answerMap["text"] == "Title" {
			if title, ok := answerMap["answer"].(string); ok {
				movie.Title = title
			}
		}

		if answerMap["text"] == "Director" {
			if director, ok := answerMap["answer"].(string); ok {
				movie.Director = director
			}
		}

		if answerMap["text"] == "Year" {
			if year, ok := answerMap["answer"].(string); ok {
				movie.Year = year
			}
		}

		if answerMap["text"] == "Notes" {
			if notes, ok := answerMap["answer"].(string); ok {
				movie.Notes = notes
			}
		}
	}

	return movie
}

func fillMovieStructMetadata(movie *Movie) *Movie {
	omdbApiURL := "https://www.omdbapi.com/?apikey=" + OMDBAPIKey + "&t=" + url.QueryEscape(movie.Title)

	res, err := http.Get(omdbApiURL)
	if err != nil {
		log.Fatal(err)
	}

	// Override the value from OMDB API to existing key in movie struct
	if err := json.NewDecoder(res.Body).Decode(movie); err != nil {
		fmt.Println("Could not decode JSON", err)
		return movie
	}

	return movie
}

func getRandomMovie() *Movie {

	allMovies := getAllMovies()

	// len(allMovies) is the total number of movies
	// pick a random one
	rand.Seed(time.Now().UnixNano())
	randKey := rand.Intn(len(allMovies))
	randomMovie := allMovies[randKey]

	randomMovieStruct := createMovieStructFromJotFormResponse(randomMovie)
	randomMovieStruct = fillMovieStructMetadata(randomMovieStruct)

	return randomMovieStruct
}

func sendMovieToChat(chatID int64, movie Movie) error {

	t := template.Must(template.ParseFiles("movie.tmpl"))
	var tpl bytes.Buffer
	err := t.Execute(&tpl, movie)
	if err != nil {
		panic(err)
	}

	// Create the request body struct
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   tpl.String(),
		ParseMode: "HTML",
	}

	// Create the JSON body from the struct
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Send a post request with your token
	telegramApiUrl := "https://api.telegram.org/bot" + TelegramToken + "/sendMessage"
	res, err := http.Post(telegramApiUrl, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return errors.New("unexpected status" + res.Status + " -> " + string(body))
	}

	return nil
}

func handleRequests() {
	router := mux.NewRouter().StrictSlash(true)

	router.HandleFunc("/", homePage)
	router.HandleFunc("/all", returnAllMovies)
	router.HandleFunc("/telegram", telegramWebhook)

	log.Fatal(http.ListenAndServeTLS(":443", CertFilePath, KeyFilePath, router))
}

func init() {

	// TODO avoid the global variables, migrate them to kind of Config!

	TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	if TelegramToken == "" {
		fmt.Println("TELEGRAM_TOKEN must be set!")
		os.Exit(1)
	}

	JotFormAPIKey = os.Getenv("JOTFORM_API_KEY")
	if JotFormAPIKey == "" {
		fmt.Println("JOTFORM_API_KEY must be set!")
		os.Exit(1)
	}

	JotFormFormID = os.Getenv("JOTFORM_FORM_ID")
	if JotFormFormID == "" {
		fmt.Println("JOTFORM_FORM_ID must be set!")
		os.Exit(1)
	}

	OMDBAPIKey = os.Getenv("OMDB_API_KEY")
	if OMDBAPIKey == "" {
		fmt.Println("OMDB_API_KEY must be set!")
		os.Exit(1)
	}

	CertFilePath = os.Getenv("CERT_FILE_PATH")
	if CertFilePath == "" {
		var path string
		path, err := filepath.Abs("cert.pem")
		if err != nil {
			fmt.Println("Error on Abs cert.pem")
		}
		CertFilePath = path
	}

	KeyFilePath = os.Getenv("KEY_FILE_PATH")
	if KeyFilePath == "" {
		var path string
		path, err := filepath.Abs("key.pem")
		if err != nil {
			fmt.Println("Error on Abs key.pem")
		}
		KeyFilePath = path
	}
}

func main() {
	fmt.Println("Rest API started...")
	handleRequests()
}
