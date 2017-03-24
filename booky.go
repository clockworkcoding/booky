package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/clockworkcoding/goodreads"
)

var config Configuration

type Configuration struct {
	GoodReadsKey    string `json:"goodReadsKey"`
	GoodReadsSecret string `json:"goodReadsSecret"`
	SlackToken      string `json:"slackToken"`
	GoodReadsHost   string `json:"goodReadsHost"`
	GoodReadsPort   string `json:"goodReadsPort"`
	SlackHost       string `json:"slackHost"`
	SlackHostHTTP   string `json:"slackHostHttp"`
	IsHTTPS         string `json:"isHTTPS"`
}

func main() {
	gr := goodreads.NewClient(config.GoodReadsKey, config.GoodReadsSecret)
	results, err := gr.GetSearch("Collapsing Empire")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	book, err := gr.GetBook(results.Search_work[0].Search_best_book.Search_id.Text)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(book.Book_title[0].Text)
	fmt.Println(book.Book_description.Text)
}

func init() {
	config = readConfig()
}

func readConfig() Configuration {

	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}

	if configuration.GoodReadsKey = os.Getenv("goodReadsKey"); configuration.GoodReadsKey != "" {
		configuration.GoodReadsHost = os.Getenv("goodReadsHost")
		configuration.GoodReadsPort = os.Getenv("goodReadsPort")
		configuration.GoodReadsSecret = os.Getenv("goodReadsSecret")
		configuration.IsHTTPS = os.Getenv("isHTTPS")
		configuration.SlackHostHTTP = os.Getenv("slackHostHTTP")
		configuration.SlackHost = os.Getenv("slackHost")
		configuration.SlackToken = os.Getenv("slackToken")
	} else {
		err := decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("error:", err)
		}
	}

	return configuration
}
