package client

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/chzyer/readline"
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

var rl *readline.Instance

func init() {
	var err error
	rl, err = readline.New("> ")
	if err != nil {
		panic(err)
	}
}

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
			if err := cfg.client.setOffline(cfg.currentUser.Token); err != nil {
				fmt.Printf("Couldn't set user offline. %v\n", err)
			}
		}
		os.Exit(0)
	}()

	defer rl.Close()

	// repl
	for {
		input := getInput("Enter command")
		if len(input) == 0 {
			continue
		}

		command := input[0]
		switch command {
		case "register":
			handleRegister(cfg)
		case "login":
			handleLogin(cfg)
		case "chat":
			if isAuthenticated(cfg) {
				handleChat(cfg, input)
			}
		case "friends":
			if isAuthenticated(cfg) {
				handleFriends(cfg, input)
			}
		case "notifications":
			if isAuthenticated(cfg) {
				handleNotifications(cfg)
			}
		case "help":
			handleHelp()
		case "exit":
			handleExit(cfg)
		case "/exit":
			handleExit(cfg)
		default:
			fmt.Println("Invalid command.")
			handleHelp()
		}
	}
}

func getInput(msg string) []string {
	if len(msg) > 0 {
		fmt.Println(msg)
	}
	line, err := rl.Readline()
	if err != nil {
		return nil
	}
	line = strings.TrimSpace(line)
	return strings.Fields(line)
}

func isAuthenticated(cfg *config) bool {
	if cfg.currentUser.Token == "" {
		fmt.Println("You must be logged in first")
		return false
	}
	return true
}

func handleRegister(cfg *config) {
	nickname, password := getRegisterDetails(cfg)
	result, err := cfg.client.register(nickname, password)
	if err != nil {
		fmt.Printf("Couldn't register. %v\n", err)
		return
	}

	fmt.Printf("Registered as %s\n", result.Nickname)
}

func handleLogin(cfg *config) {
	nickname, password := getLoginDetails(cfg)
	result, err := cfg.client.login(nickname, password)
	if err != nil {
		fmt.Printf("Couldn't login. %v\n", err)
		return
	}

	cfg.currentUser.Nickname = result.Nickname
	cfg.currentUser.Token = result.Token

	if err := cfg.client.setOnline(cfg.currentUser.Token); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("==============================")
	fmt.Printf("Login successful as %s\n", result.Nickname)

	handleNotifications(cfg)
}

func handleChat(cfg *config, input []string) {
	if len(input) < 2 {
		fmt.Println("==============================")
		fmt.Println("Invalid input. Usage:")
		fmt.Println()
		fmt.Printf("  %-30s %s\n", "chat <nickname>", "open chat with friend")
		fmt.Println()
		fmt.Println("==============================")
		return
	}
	targetNickname := input[1]
	result, err := cfg.client.startChat(targetNickname, cfg.currentUser.Token)
	if err != nil {
		fmt.Printf("Couldn't open chat. %v\n", err)
		return
	}

	if err := cfg.client.connectToChat(result.ChatID, cfg.currentUser.Token); err != nil {
		fmt.Printf("Couldn't connect to chat. %v\n", err)
		return
	}
}

func handleFriends(cfg *config, input []string) {
	if len(input) < 2 {
		displayFriendsHelp()
		return
	}

	// args
	switch input[1] {
	case "--list":
		displayFriendsList(cfg)
		return
	case "--delete":
		if len(input) < 3 {
			fmt.Println("==============================")
			fmt.Println("Invalid input")
			fmt.Printf("  %-30s %s\n", "friends --delete <nickname>", "remove friend")
			fmt.Println("==============================")
			return
		}

		targetName := input[2]

		err := cfg.client.deleteFriendship(targetName, cfg.currentUser.Token)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Printf("You have successfully removed %s from your friends list.\n", targetName)
		return
	case "--help":
		displayFriendsHelp()
		return
	default:
		targetNickname := input[1]

		result, err := cfg.client.sendFriendRequest(targetNickname, cfg.currentUser.Token)
		if err != nil {
			fmt.Println(err)
			return
		}

		switch result.Status {
		case "created":
			fmt.Printf("Friend request successfully sent to %s.\n", targetNickname)
			return
		case "accepted":
			fmt.Printf("You have successfully added %s as a friend.\n", targetNickname)
			return
		case "sent":
			fmt.Printf("Friend request to %s already sent.\n", targetNickname)
			return
		case "friends":
			fmt.Printf("You and %s are already friends.\n", targetNickname)
			return
		}
	}
}

func handleNotifications(cfg *config) {
	notifications, err := cfg.client.getNotifications(cfg.currentUser.Token)
	if err != nil {
		fmt.Printf("Couldn't get notifications. %v\n", err)
		return
	}

	if len(notifications.UnreadMessages) == 0 && len(notifications.FriendRequests) == 0 {
		fmt.Println()
		fmt.Println("You have no notifications")
		fmt.Println("==============================")
		return
	}

	if len(notifications.UnreadMessages) > 0 {
		fmt.Println()
		for _, msg := range notifications.UnreadMessages {
			fmt.Printf("You have %d unread messages from %s\n", msg.Count, msg.Nickname)
		}
	}

	if len(notifications.FriendRequests) > 0 {
		fmt.Println()
		for _, f := range notifications.FriendRequests {
			if f.SenderNickname == cfg.currentUser.Nickname {
				fmt.Printf("Friend request to %s is pending\n", f.ReceiverNickname)
			} else {
				fmt.Printf("Friend request from %s\n", f.SenderNickname)
			}
		}
	}
	fmt.Println("==============================")
}

func handleExit(cfg *config) {
	fmt.Println()
	fmt.Println("Closing the messenger... Goodbye!")
	fmt.Println()
	if cfg.currentUser.Token != "" {
		if err := cfg.client.setOffline(cfg.currentUser.Token); err != nil {
			fmt.Printf("Couldn't set user offline: %s\n", err)
		}
	}
	if err := rl.Close(); err != nil {
		fmt.Printf("Couldn't close readline: %s\n", err)
	}
	os.Exit(0)
}

func handleHelp() {
	fmt.Println("==============================")
	fmt.Println()
	fmt.Printf("  %-30s %s\n", "register", "create a new account")
	fmt.Printf("  %-30s %s\n", "login", "sign in with your nickname and password")
	fmt.Printf("  %-30s %s\n", "chat <nickname>", "open a chat with a friend")
	fmt.Printf("  %-30s %s\n", "friends <nickname>", "send a friend request (friends --help for more)")
	fmt.Printf("  %-30s %s\n", "notifications", "view unread messages and friend requests")
	fmt.Printf("  %-30s %s\n", "help", "show available commands")
	fmt.Printf("  %-30s %s\n", "exit", "quit the application")
	fmt.Println()
	fmt.Println("==============================")
}

func displayFriendsHelp() {
	fmt.Println("==============================")
	fmt.Println("Invalid input. Usage:")
	fmt.Println()
	fmt.Printf("  %-30s %s\n", "friends <nickname>", "send friend request")
	fmt.Printf("  %-30s %s\n", "friends --list", "show all friends")
	fmt.Printf("  %-30s %s\n", "friends --delete <nickname>", "remove friend")
	fmt.Printf("  %-30s %s\n", "friends --help", "show available commands")
	fmt.Println()
	fmt.Println("==============================")
}

func displayFriendsList(cfg *config) {
	friends, err := cfg.client.getFriends(cfg.currentUser.Token)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("==============================")
	fmt.Println("Your friends:")
	count := 1
	for _, f := range friends {
		status := "offline"
		if f.Online {
			status = "online"
		}
		fmt.Printf("%d.%-15s [%s]\n", count, f.Nickname, status)
		count += 1
	}
	fmt.Println("==============================")
}

func getRegisterDetails(cfg *config) (string, string) {
	for {
		fmt.Println()
		nickname := getInput("Enter a nickname")
		if len(nickname) == 0 {
			fmt.Println("you have not entered a nickname")
			continue
		}
		if nickname[0] == "/exit" || nickname[0] == "exit" {
			handleExit(cfg)
		}
		if len(nickname[0]) < 4 && len(nickname[0]) > 20 {
			fmt.Println("nickname must be from 4 to 20 characters in length")
			continue
		}

		fmt.Println()
		password := getInput("Enter a password")
		if len(password) == 0 {
			fmt.Println("you have not entered a password")
			continue
		}
		if password[0] == "/exit" || password[0] == "exit" {
			handleExit(cfg)
		}
		if len(password[0]) < 6 {
			fmt.Println("password must be longer than 6 characters")
			continue
		}

		return nickname[0], password[0]
	}
}

func getLoginDetails(cfg *config) (string, string) {
	for {
		fmt.Println()
		nickname := getInput("Enter a nickname")
		if len(nickname) == 0 {
			fmt.Println("you have not entered a nickname")
			continue
		}
		if nickname[0] == "/exit" || nickname[0] == "exit" {
			handleExit(cfg)
		}

		fmt.Println()
		password := getInput("Enter a password")
		if len(password) == 0 {
			fmt.Println("you have not entered a password")
			continue
		}
		if password[0] == "/exit" || password[0] == "exit" {
			handleExit(cfg)
		}

		return nickname[0], password[0]
	}
}
