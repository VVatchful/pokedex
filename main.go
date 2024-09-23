package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Println("You found someones pokedex!")

	commands := map[string]func(){
		"help": printHelp,
		"exit": exitREPL,
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("pokedex > ")

		scanner.Scan()
		input := strings.TrimSpace(scanner.Text())

		if cmd, found := commands[input]; found {
			cmd()
		} else {
			fmt.Println("Unknown command. Type 'help' to see available commands.")
		}
	}
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help  - Prints this help message.")
	fmt.Println("  exit  - Exits the REPL.")
}

func exitREPL() {
	fmt.Println("Goodbye!")
	os.Exit(0) }

