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

type CurrentUser struct {
	Nickname string
	Token    string
}

type config struct {
	client      Client
	currentUser CurrentUser
}

var scanner = bufio.NewScanner(os.Stdin)

func Run() {
	timeout := 5 * time.Second
	client := Client{httpClient: http.Client{Timeout: timeout}}
	currentUser := CurrentUser{}

	config := &config{
		client:      client,
		currentUser: currentUser,
	}

	for {
		input := getInput("Enter command")
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "register":
			nickname, password := getLoginDetails()
			result, err := config.client.Register(nickname, password)
			if err != nil {
				fmt.Printf("Couldn't register. %v\n", err)
				continue
			}

			fmt.Printf("Registered as %s\n", result.Nickname)
			continue

		case "login":
			nickname, password := getLoginDetails()
			result, err := config.client.Login(nickname, password)
			if err != nil {
				fmt.Printf("Couldn't login. %v\n", err)
				continue
			}

			config.currentUser.Nickname = result.Nickname
			config.currentUser.Token = result.Token

			fmt.Printf("Login successful as %s\n", result.Nickname)
			continue
		}
	}
}

func getInput(msg string) []string {
	if len(msg) > 0 {
		fmt.Println(msg)
	}
	fmt.Print("> ")
	if !scanner.Scan() {
		return nil
	}
	line := strings.TrimSpace(scanner.Text())
	return strings.Fields(line)
}

func getLoginDetails() (string, string) {
	for {
		nickname := getInput("Enter nickname")
		if len(nickname) == 0 {
			fmt.Println("you have not entered nickname")
			continue
		}
		password := getInput("Enter password")
		if len(password) == 0 {
			fmt.Println("you have not entered password")
			continue
		}

		return nickname[0], password[0]
	}
}
