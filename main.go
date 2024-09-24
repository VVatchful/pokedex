package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

)

type Config struct {
  nextURl     string
  previousURL string
}

type LocationArea struct {
  Name string `json:"name"`
}

type LocationAreaResponse struct {
  Results []LocationArea  `json:"results"`
  Next     *string        `json:"next"`
  Previous *string        `json:"previous"`
}

func main() {
	fmt.Println("You found someones pokedex!")

  config := &Config{}

  commands := map[string]func(*Config){
    "help":  func(config *Config) { printHelp() },
		"exit":  func(config *Config) { exitREPL() },
		"map":   mapCommand,
		"mapb":  mapBackCommand,
  }
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("pokedex > ")

		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if cmd, found := commands[input]; found {
			cmd(config)
		} else {
			fmt.Println("Unknown command. Type 'help' to see a list available commands.")
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help  - Prints this help message.")
	fmt.Println("  exit  - Exits the pokedex.")
	fmt.Println("  map   - Displays the next 20 location areas.")
	fmt.Println("  mapb  - Displays the previous 20 location areas.")
}

func exitREPL() {
	fmt.Println("Goodbye!")
	os.Exit(0)
}

func fetchLocationAreas(url string) (*LocationAreaResponse, error) {
  resp, err := http.Get(url)
  if err != nil {
    return nil, fmt.Errorf("Failed to fetch data: %w", err)

  }
  defer resp.Body.Close()

  body, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, fmt.Errorf("Failed to read response body: %w", err)
  }

  var locationAreas LocationAreaResponse
  err = json.Unmarshal(body, &locationAreas)
  if err != nil {
    return nil, fmt.Errorf("Failed to parse JSON: %w", err)
  }

  return &locationAreas, nil
}

func mapCommand(config *Config) {
  url := config.nextURl
  if url == "" {
    url = "https://pokeapi.co/api/v2/location-area/"
  }

  locationAreas, err := fetchLocationAreas(url)
  if err != nil {
    fmt.Println("Error:",err)
    return
  }

  for _, area := range locationAreas.Results {
    fmt.Println(area.Name)
  }

  if locationAreas.Next != nil {
    config.nextURl = *locationAreas.Next
  }
  if locationAreas.Previous != nil {
    config.previousURL = *locationAreas.Previous
  }
}

func mapBackCommand(config *Config) {
    if config.previousURL == "" {
      fmt.Println("You are already on the first page, cannot go back any further.")
      return
    }

    locationAreas, err := fetchLocationAreas(config.previousURL)
    if err != nil {
      fmt.Println("Error", err)
      return
    }

    for _, area := range locationAreas.Results {
      fmt.Println(area.Name)
    }

    if locationAreas.Next != nil {
      config.nextURl = *locationAreas.Next
    }
    if locationAreas.Previous != nil {
      config.previousURL = *locationAreas.Previous
    }
}

