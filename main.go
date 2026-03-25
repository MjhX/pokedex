// main.go
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
    "github.com/MjhX/pokedex/internal/pokecache"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*cliConfig, []string, *map[string]Pokemon) error
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

type ExploreResponse struct {
	Encounters []ExploreEncounters  `json:"pokemon_encounters"`
}

type ExploreEncounters struct {
    Pokemon NameURL `json:"pokemon"`
}

type NameURL struct {
    Name  string `json:"name"`
	URL  *string `json:"url"`
}

type Pokemon struct {
	BaseEXP int64  `json:"base_experience"`
	Name    string `json:"name"`
	Height  int    `json:"height"`
	Weight  int    `json:"weight"`
	Stats   []Stat `json:"stats"`
	Types   []Type `json:"types"`
}

type cliConfig struct {
	next     *string
	previous *string
	cache     pokecache.Cache
}

type Stat struct {
    BaseStat  int `json:"base_stat"`
    Stat      NameURL `json:"stat"`
}

type Type struct {
	Type NameURL `json:"type"`
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
		"explore": {
			name:        "explore",
			description: "Shows list of Pokémon in an area",
			callback:    commandExplore,
		},
		"catch": {
			name:        "catch",
			description: "Attempts to catch a Pokemon",
			callback:    commandCatch,
		},
		"inspect": {
			name:        "inspect",
			description: "See details of a caught Pokemon",
			callback:    commandInspect,
		},
		"pokedex": {
			name:        "pokedex",
			description: "See list of caught Pokemon",
			callback:    commandPokedex,
		},
	}
}

func cleanInput(text string) []string {
	s := strings.Fields(strings.ToLower(text))
	return s
}

func commandExit(_ *cliConfig, args []string, dex *map[string]Pokemon) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return nil
}

func commandHelp(_ *cliConfig, args []string, dex *map[string]Pokemon) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println("")
	for k, v := range getCommands() {
		fmt.Printf("%s: %s\n", k, v.description)
	}
	return nil
}

func commandMap(c *cliConfig, args []string, dex *map[string]Pokemon) error {
	if c.next == nil {
		fmt.Println("you're on the last page")
		return nil
	}
	return readOutMap(*(c.next), c)
}

func commandMapb(c *cliConfig, args []string, dex *map[string]Pokemon) error {
	if c.previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}
	return readOutMap(*(c.previous), c)
}

func commandExplore(c *cliConfig, args []string, dex *map[string]Pokemon) error {
	if len(args) == 0 {
		fmt.Println("Include an area")
		return nil
	}
	body, e := getBody("https://pokeapi.co/api/v2/location-area/"+args[0], c)
	if e != nil {
		fmt.Println(args[0]+" is not a valid location")
		return nil
	}
	fmt.Println("Exploring "+args[0]+"...")

	var ER ExploreResponse
    e3 := json.Unmarshal(body, &ER)
	if e3 != nil {
		log.Fatal(e3)
	}
	fmt.Println("Found Pokemon:")
	for _, a := range ER.Encounters {
		fmt.Println(" - "+a.Pokemon.Name)
	}
	return nil
}

func commandCatch(c *cliConfig, args []string, dex *map[string]Pokemon) error {
	if len(args) == 0 {
		fmt.Println("Include a Pokemon")
		return nil
	}
	body, e := getBody("https://pokeapi.co/api/v2/pokemon/"+args[0], c)
	if e != nil {
		fmt.Println(args[0]+" is not a valid location")
		return nil
	}
	var PKMN Pokemon
	er := json.Unmarshal(body, &PKMN)
	if er != nil {
		log.Fatal(er)
	}
	fmt.Println("Throwing a Pokeball at "+PKMN.Name+"...")
    if rand.Int63n(1000+PKMN.BaseEXP) > PKMN.BaseEXP {
		var P NameURL
		ep := json.Unmarshal(body, &P)
		if ep != nil {
			log.Fatal(ep)
		}
        fmt.Println(PKMN.Name+" was caught!")
		(*dex)[PKMN.Name] = PKMN
		fmt.Println("You may now inspect it with the inspect command.")
	} else {
        fmt.Println(PKMN.Name+" escaped!")
	}
	return nil
}

func commandInspect(c *cliConfig, args []string, dex *map[string]Pokemon) error {
    if len(args) == 0 {
		fmt.Println("Include a Pokemon")
		return nil
	}
	pkmn, ok := (*dex)[args[0]]
	if ok == false {
		fmt.Println("you have not caught that pokemon")
		return nil
	}
	fmt.Println("Name: "+pkmn.Name)
	fmt.Printf("Height: %d\n",pkmn.Height)
	fmt.Printf("Weight: %d\n",pkmn.Weight)
	fmt.Println("Stats:")
	for _, stat := range pkmn.Stats {
	    fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, t := range pkmn.Types {
	    fmt.Printf("  - %s\n", t.Type.Name)
	}
	return nil
}

func commandPokedex(c *cliConfig, args []string, dex *map[string]Pokemon) error {
	fmt.Println("Your Pokedex:")
	for _, pkmn := range (*dex) {
        fmt.Println("  - "+pkmn.Name)
	}
	return nil
}

func getBody(url string, c *cliConfig) ([]byte, error) {
	var body []byte
	val, ok := c.cache.Get(url)
	if ok {
        body = val
	} else {
		res, e1 := http.Get(url)
		if e1 != nil {
			log.Fatal(e1)
		}
		body2, e2 := io.ReadAll(res.Body)
		defer res.Body.Close()
		if e2 != nil {
			log.Fatal(e2)
		}
		c.cache.Add(url, body2)
		body = body2
	}
	return body, nil
}

func readOutMap(url string, c *cliConfig) error {
	body, e1 := getBody(url, c)
	if e1 != nil {
		log.Fatal(e1)
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
	CONFIG.cache = pokecache.NewCache(5*time.Minute)
	DEX := make(map[string]Pokemon)

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
			v.callback(&CONFIG, scanText[1:], &DEX)
		} else {
			fmt.Printf("Your command was: %s\n", scanText[0])
		}
	}
}
