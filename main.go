package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

  "github.com/VVatchful/pokedex/pokecache"
)

type LocationAreaResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
	Next    *string `json:"next"`
	Previous *string `json:"previous"`
}

type PokemonAreaResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

var cache = pokecache.NewCache(5 * time.Minute)
var nextURL *string
var prevURL *string

func main() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("pokedex > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		parts := strings.SplitN(input, " ", 2)
		command := parts[0]
		var argument string
		if len(parts) > 1 {
			argument = parts[1]
		}

		switch command {
		case "help":
			printHelp()
		case "exit":
			fmt.Println("Exiting...")
			return
		case "map":
			handleMap(false)
		case "mapb":
			handleMap(true)
		case "explore":
			if argument == "" {
				fmt.Println("Usage: explore <location_area>")
			} else {
				handleExplore(argument)
			}
		default:
			fmt.Println("Unknown command. Type 'help' for the list of available commands.")
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("help    - Show this help message")
	fmt.Println("exit    - Exit the program")
	fmt.Println("map     - Show 20 location areas")
	fmt.Println("mapb    - Go back 20 location areas")
	fmt.Println("explore - Explore a specific location area (usage: explore <area_name>)")
}

func handleMap(isBack bool) {
	var url string
	if isBack {
		if prevURL == nil {
			fmt.Println("No previous location areas available.")
			return
		}
		url = *prevURL
	} else {
		if nextURL == nil {
			url = "https://pokeapi.co/api/v2/location-area/"
		} else {
			url = *nextURL
		}
	}

	if data, found := cache.Get(url); found {
		fmt.Println("Cached data found:")
		parseAndDisplayLocations(data)
		return
	}

	data, err := fetchData(url)
	if err != nil {
		fmt.Println("Error fetching location areas:", err)
		return
	}

	cache.Add(url, data)

	parseAndDisplayLocations(data)
}

func handleExplore(areaName string) {
	// Check if the location area data is cached
	if data, found := cache.Get(areaName); found {
		fmt.Println("Cached data found:")
		parseAndDisplayPokemon(data)
		return
	}

	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", areaName)
  fmt.Println("Fetching from URL:", url)
	data, err := fetchData(url)
	if err != nil {
		fmt.Println("Error fetching location area data:", err)
		return
	}

	cache.Add(areaName, data)

	parseAndDisplayPokemon(data)
}

func fetchData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func parseAndDisplayLocations(data []byte) {
	var locationAreas LocationAreaResponse
	err := json.Unmarshal(data, &locationAreas)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Println("Location areas:")
	for _, location := range locationAreas.Results {
		fmt.Println(location.Name)
	}

	nextURL = locationAreas.Next
	prevURL = locationAreas.Previous
}

func parseAndDisplayPokemon(data []byte) {
	var pokemonArea PokemonAreaResponse
	err := json.Unmarshal(data, &pokemonArea)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return
	}

	fmt.Println("Pokémon found in this area:")
	for _, encounter := range pokemonArea.PokemonEncounters {
		fmt.Println(encounter.Pokemon.Name)
	}
}
