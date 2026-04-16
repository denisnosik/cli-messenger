package client

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "http://localhost:8080"

type Client struct {
	httpClient http.Client
}

type config struct {
	client Client
}

func getInput(msg string) []string {
	if len(msg) > 0 {
		fmt.Println(msg)
	}
	fmt.Print("> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func Run() {
	timeout := 5 * time.Second
	client := Client{httpClient: http.Client{Timeout: timeout}}

	config := &config{
		client: client,
	}

	for {
		input := getInput("Enter command")
		if len(input) == 0 {
			continue
		}

		switch input[0] {
		case "registration":
			for {
				nickname := getInput("Enter nickname")
				if len(nickname) == 0 {
					fmt.Println("Invalid nickname")
					continue
				}

				password := getInput("Enter password")
				if len(password) == 0 {
					fmt.Println("Invalid password")
					continue
				}

				user, err := config.client.Register(nickname[0], password[0])
				if err != nil {
					fmt.Println("Error:", err)
					continue
				}

				fmt.Printf("Registered as %s (id: %s)\n", user.Nickname, user.ID)
				break
			}

		}
	}
}
