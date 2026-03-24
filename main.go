// main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
    "github.com/MjhX/pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*cliConfig) error
}

type Location struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type LocationResponse struct {
	Count    int        `json:"count"`
	Next     *string    `json:"next"`
	Previous *string    `json:"previous"`
	Results  []Location `json:"results"`
}

type cliConfig struct {
	next     *string
	previous *string
	cache     Cache
}

func getCommands() map[string]cliCommand {
	return map[string]cliCommand{
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"map": {
			name:        "map",
			description: "Lists next 20 world locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Lists previous 20 world locations",
			callback:    commandMapb,
		},
	}
}

func cleanInput(text string) []string {
	s := strings.Fields(strings.ToLower(text))
	return s
}

func commandExit(_ *cliConfig) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *cliConfig) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for k, v := range getCommands() {
		fmt.Printf("%s: %s\n", k, v.description)
	}
	return nil
}

func commandMap(c *cliConfig) error {
	if c.next == nil {
		fmt.Println("you're on the last page")
	}
	return readOutMap(*(c.next), c)
}

func commandMapb(c *cliConfig) error {
	if c.previous == nil {
		fmt.Println("you're on the first page")
	}
	return readOutMap(*(c.previous), c)
}

func readOutMap(url string, c *cliConfig) error {
	var body []byte
	val, ok := c.cache.entries[url]
	if ok {
      body = val.val
	} else {
		res, e1 := http.Get(url)
		if e1 != nil {
			log.Fatal(e1)
		}
		body2, e2 := io.ReadAll(res.Body)
		if e2 != nil {
			log.Fatal(e2)
		}
		defer res.Body.Close()
		body = body2
	}
	
	var LR LocationResponse
	e3 := json.Unmarshal(body, &LR)
	if e3 != nil {
		log.Fatal(e3)
	}
	c.next, c.previous = LR.Next, LR.Previous
	for _, a := range LR.Results {
		fmt.Println(a.Name)
	}
	return nil
}

func main() {
	var CONFIG cliConfig
	FIRST_URL := "https://pokeapi.co/api/v2/location-area/"
	CONFIG.next = &FIRST_URL

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		canScan := scanner.Scan()
		if canScan == false {
			break
		}
		scanText := cleanInput(scanner.Text())
		v, ok := getCommands()[scanText[0]]
		if ok {
			v.callback(&CONFIG)
		} else {
			fmt.Printf("Your command was: %s\n", scanText[0])
		}
	}
}
