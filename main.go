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
  "math/rand"

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

type Pokemon struct {
  Name           string `json:"name"`
  BaseExperience int    `json:"base_experience"`
  Height         int    `json:"height"`
  Weight         int    `json:"weight"`
  Stats          []Stat `json:"stats"`
  Types          []TypeElement `json:"types"`
}

type Stat struct {
  BaseStat int `json:"base_stat"`
  Stat     struct {
    Name string
  } `json:"type"`
}

type TypeElement struct {
  Type struct {
    Name string `json:"name"`
  } `json:"type"`
}


var pokedex = make(map[string]Pokemon)
var cache = pokecache.NewCache(5 * time.Minute)
var nextURL *string
var prevURL *string

func main() {
	reader := bufio.NewReader(os.Stdin)

   commands := map[string]func(args []string){
        "inspect": func(args []string) {
            if len(args) < 1 {
                fmt.Println("Usage: inspect <pokemon_name>")
                return
            }
            handleInspect(args[0])
        },
        "catch": func(args []string) {
            if len(args) < 1 {
                fmt.Println("Usage: catch <pokemon_name>")
                return
            }
            handleCatch(args[0])
        },
        "pokedex": func(args []string) {
            if len(args) != 0 {
              fmt.Println("Usage: pokedex")
              return
          }
          handlePokedex()
        },

    }


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
    case "catch":
      catchArgs := strings.Fields(argument)
      commands["catch"](catchArgs)
    case "inspect":
      inspectArgs := strings.Fields(argument)
      commands["inspect"](inspectArgs)
    case "pokedex":
      pokedexArgs := strings.Fields(argument)
      commands["pokedex"](pokedexArgs)
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
    fmt.Println("catch   - Attempts to catch the pokemon")
    fmt.Println("inspect - inspects the caught pokemon")
    fmt.Println("pokedex - displays all the pokemon you caught")
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

func handleCatch(pokemonName string) {
    if data, found := cache.Get(pokemonName); found {
        attemptToCatchPokemon(data)
        return
    }

    url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", pokemonName)
    data, err := fetchData(url)
    if err != nil {
        fmt.Println("Error fetching Pokémon data:", err)
        return
    }

    cache.Add(pokemonName, data)

    attemptToCatchPokemon(data)
}

func attemptToCatchPokemon(data []byte) {
    var pokemon Pokemon
    if err := json.Unmarshal(data, &pokemon); err != nil {
        fmt.Println("Error parsing Pokémon data:", err)
        return
    }

    catchChance := rand.Float64() < (100.0 / (float64(pokemon.BaseExperience) + 50.0))

    if catchChance {
        fmt.Printf("Success! You caught %s.\n", pokemon.Name)
        addPokemonToPokedex(pokemon)
    } else {
        fmt.Printf("%s escaped!\n", pokemon.Name)
    }
}

func addPokemonToPokedex(pokemon Pokemon) {
    if _, exists := pokedex[pokemon.Name]; exists {
        fmt.Printf("%s is already in your Pokedex!\n", pokemon.Name)
    } else {
        pokedex[pokemon.Name] = pokemon
        fmt.Printf("%s has been added to your Pokedex.\n", pokemon.Name)
    }
}

func handleInspect(pokemonName string) {
  if pokemon, exists := pokedex[pokemonName]; exists {
    fmt.Println("Name: %s\n", pokemon.Name)
    fmt.Println("Height: %d decimetres\n", pokemon.Height)
    fmt.Println("Weight: %d hectograms\n", pokemon.Weight)

    fmt.Println("Stats:")
    for _, stat := range pokemon.Stats {
      fmt.Printf("  %s: %d\n", stat.Stat.Name, stat.BaseStat)
    }

    fmt.Println("Types:")
    for _, pokemonType := range pokemon.Types {
      fmt.Printf("  %s\n", pokemonType.Type.Name)
    }
  } else {
    fmt.Printf("You haven't caught %s yet!\n", pokemonName)
  }
}

func handlePokedex() {
  if len(pokedex) == 0 {
    fmt.Println("You have not caught any pokemon yet!")
    return
  }

  fmt.Println("Your Pokedex contains the following pokemon:")
  for name := range pokedex {
    fmt.Println("-", name)
  }
}
