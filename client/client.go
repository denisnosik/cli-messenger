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

		case "chat":
			targetNickname := getChatTargetNickname()
			result, err := config.client.StartChat(targetNickname, config.currentUser)
			if err != nil {
				fmt.Printf("Couldn't open chat. %v\n", err)
				continue
			}
			fmt.Printf("Star chat with id: %s\n", result.ChatID)

			client.ConnectToChat(result.ChatID, config.currentUser.Token)

		case "exit":
			fmt.Println("Closing the messenger... Goodbye!")
			os.Exit(0)

		default:
			fmt.Println("Invalid command.")
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
		nickname := getInput("Enter a nickname")
		if len(nickname) == 0 {
			fmt.Println("you have not entered a nickname")
			continue
		}
		password := getInput("Enter a password")
		if len(password) == 0 {
			fmt.Println("you have not entered a password")
			continue
		}

		return nickname[0], password[0]
	}
}

func getChatTargetNickname() string {
	for {
		targetNickname := getInput("Enter a nickname")
		if len(targetNickname) == 0 {
			fmt.Println("you have not entered a nickname")
			continue
		}

		return targetNickname[0]
	}
}
