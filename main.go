package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Movie struct {
	Title string `json:"Title"`
	Year int16 `json:"year"`
	Director string `json:"director"`
}

// global Movies array to simulate a database, for now
var Movies []Movie

// SSL certificate path
var CertFilePath string
// SSL private key path
var KeyFilePath string
// Telegram token
var TelegramToken string

func homePage(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Welcome to the Homepage!")
	fmt.Println("In: homePage")
}

func returnAllMovies(w http.ResponseWriter, r *http.Request){
	fmt.Println("In: returnAllMovies")
	json.NewEncoder(w).Encode(Movies)
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

	// Say hello back to user with name
	if err := sayHello(body.Message.Chat.ID, body.Message.User.Name); err != nil {
		fmt.Println("Error in sending reply:", err)
		return
	}

	// log a confirmation message if the message is sent successfully
	fmt.Println("Reply sent!")
}

func sayHello(chatID int64, name string) error {
	// Create the request body struct
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   "Hello, " + name + "!",
	}

	// Create the JSON body from the struct
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Send a post request with your token
	url := "https://api.telegram.org/bot" + TelegramToken + "/sendMessage"
	res, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
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
	Movies = []Movie{
		{Title: "The Royal Tenenbaums", Year: 2001, Director: "Wes Anderson"},
		{Title: "The Italian Job", Year: 2003, Director: "F. Gary Gray"},
	}
	handleRequests()
}
