package client

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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

	cfg := &config{
		client:      client,
		currentUser: currentUser,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if cfg.currentUser.Token != "" {
			cfg.client.SetOffline(cfg.currentUser.Token)
		}
		os.Exit(0)
	}()

	// repl
	for {
		input := getInput("Enter command")
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "register":
			nickname, password := getLoginDetails()
			result, err := cfg.client.Register(nickname, password)
			if err != nil {
				fmt.Printf("Couldn't register. %v\n", err)
				continue
			}

			fmt.Printf("Registered as %s\n", result.Nickname)
			continue

		case "login":
			nickname, password := getLoginDetails()
			result, err := cfg.client.Login(nickname, password)
			if err != nil {
				fmt.Printf("Couldn't login. %v\n", err)
				continue
			}

			cfg.currentUser.Nickname = result.Nickname
			cfg.currentUser.Token = result.Token

			err = cfg.client.SetOnline(cfg.currentUser.Token)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Printf("Login successful as %s\n", result.Nickname)

			displayNotifications(cfg)
			continue

		case "chat":
			targetNickname := getChatTargetNickname()
			result, err := cfg.client.StartChat(targetNickname, cfg.currentUser.Token)
			if err != nil {
				fmt.Printf("Couldn't open chat. %v\n", err)
				continue
			}

			client.ConnectToChat(result.ChatID, cfg.currentUser.Token)

		case "friends":
			if len(input) < 2 {
				fmt.Println("Use command + nickname or command + --list")
				continue
			}

			if input[1] == "--list" {
				displayFriends(cfg)
				continue
			}

			targetNickname := input[1]
			result, err := cfg.client.handlerFriends(targetNickname, cfg.currentUser.Token)
			if err != nil {
				fmt.Println(err)
				continue
			}

			switch result.Status {
			case "created":
				fmt.Printf("Friend request successfully sent to %s.\n", targetNickname)
				continue
			case "accepted":
				fmt.Printf("You have successfully added %s as a friend.\n", targetNickname)
				continue
			case "sent":
				fmt.Printf("Friend request to %s already sent.\n", targetNickname)
				continue
			case "friends":
				fmt.Printf("You and %s are already friends.\n", targetNickname)
				continue
			}

		case "notifications":
			displayNotifications(cfg)
			continue

		case "exit":
			fmt.Println("Closing the messenger... Goodbye!")
			if cfg.currentUser.Token != "" {
				cfg.client.SetOffline(cfg.currentUser.Token)
			}
			os.Exit(0)

		default:
			fmt.Println("Invalid command.")
			continue
		}
	}
}

func displayNotifications(cfg *config) {
	notifications, err := cfg.client.GetNotifications(cfg.currentUser.Token)
	if err != nil {
		fmt.Printf("Couldn't get notifications. %v\n", err)
		return
	}

	if len(notifications.UnreadMessages) == 0 && len(notifications.FriendRequests) == 0 {
		fmt.Println("You have no notifications")
		return
	}

	if len(notifications.UnreadMessages) > 0 {
		for _, msg := range notifications.UnreadMessages {
			fmt.Printf("You have %d unread messages from %s\n", msg.Count, msg.Nickname)
		}
	}

	if len(notifications.FriendRequests) > 0 {
		for _, f := range notifications.FriendRequests {
			if f.SenderNickname == cfg.currentUser.Nickname {
				fmt.Printf("Friend request to %s is pending\n", f.ReceiverNickname)
			} else {
				fmt.Printf("Friend request from %s\n", f.SenderNickname)
			}
		}
	}
}

func displayFriends(cfg *config) {
	friends, err := cfg.client.handlerGetFriends(cfg.currentUser.Token)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	fmt.Println("Your friends:")
	count := 1
	for _, f := range friends {
		status := "offline"
		if f.Online {
			status = "online"
		}
		fmt.Printf("%d.%s [%s]\n", count, f.Nickname, status)
		count += 1
	}
	fmt.Println("------------")
	fmt.Println()
}

func getInput(msg string) []string {
	if len(msg) > 0 {
		fmt.Println(msg)
	}
	fmt.Print("> ")
	scanned := scanner.Scan()
	if !scanned {
		return nil
	}
	line := scanner.Text()
	line = strings.TrimSpace(line)
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
