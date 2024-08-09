package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var DefaultServerUrl = "http://localhost:8000"

var rootCmd = &cobra.Command{
	Use:   "goshort [url]",
	Short: "GoShort! CLI",
	Long:  `GoShort! CLI client is here to help you quickly shorten URLs from your terminal!`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Fatal("No URL provided!")
		} else {
			urlToShorten := args[0]
			handleURLShortening(urlToShorten)
		}
	},
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to the given GoShort! server",
	Run: func(cmd *cobra.Command, args []string) {
		var serverUrl string
		if len(args) == 0 {
			log.Println("No server provided! Using default server: " + DefaultServerUrl)
			serverUrl = DefaultServerUrl
		} else {
			serverUrl = args[0]
		}
		client := &http.Client{
			Timeout: 5 * time.Second,
		}
		log.Println("Checking if server is online...")
		resp, err := client.Get(serverUrl)
		if err != nil {
			log.Fatal("Failed to connect to server: " + err.Error())
		}
		if resp.StatusCode != 200 {
			log.Fatal("Server is not online! Are you sure the server is running/reachable?")
		}
		log.Println("Server is online!")
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal("Failed to close response body: " + err.Error())
			}
		}(resp.Body)

		var email string
		if err := survey.AskOne(&survey.Input{Message: "Please enter your email address:"}, &email); err != nil {
			log.Fatal("Failed to read input: " + err.Error())
		}

		var password string
		if err := survey.AskOne(&survey.Password{Message: "Please enter your password:"}, &password); err != nil {
			log.Fatal("Failed to read input: " + err.Error())
		}

		loginRequestBody := map[string]string{
			"email":    email,
			"password": password,
		}

		loginRequestBodyJson, err := json.Marshal(loginRequestBody)
		if err != nil {
			log.Fatal("Failed to marshal login request body: " + err.Error())
		}

		loginResp, err := client.Post(
			serverUrl+"/api/v1/user/login",
			"application/json",
			bytes.NewBuffer(loginRequestBodyJson),
		)
		if err != nil {
			log.Fatal("Failed to login: " + err.Error())
		}

		if loginResp.StatusCode != 200 {
			log.Fatal("Failed to login: " + loginResp.Status)
		}

		type User struct {
			Email    string `json:"email"`
			ID       int    `json:"id"`
			Username string `json:"username"`
		}

		type LoginResponse struct {
			Message string `json:"message"`
			Session string `json:"session"`
			Success bool   `json:"success"`
			User    User   `json:"user"`
		}

		var loginRespData LoginResponse
		if err := json.NewDecoder(loginResp.Body).Decode(&loginRespData); err != nil {
			log.Fatal("Failed to decode login response body: " + err.Error())
		}

		if !loginRespData.Success {
			log.Fatal("Login failed: " + loginRespData.Message)
		}

		log.Println("\nLogin successful! Your session key is: " + loginRespData.Session)

		// Store the session key in the user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get user home directory: " + err.Error())
		}

		goshortDir := filepath.Join(homeDir, ".goshort")
		if err := os.MkdirAll(goshortDir, os.ModePerm); err != nil {
			log.Fatal("Failed to create .goshort directory: " + err.Error())
		}

		sessionFilePath := filepath.Join(goshortDir, "session.txt")
		if err := os.WriteFile(sessionFilePath, []byte(loginRespData.Session), 0600); err != nil {
			log.Fatal("Failed to write session file: " + err.Error())
		}

		log.Println("Session key saved to:", sessionFilePath)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal("Failed to close response body: " + err.Error())
			}
		}(loginResp.Body)
	},
}

func main() {
	rootCmd.AddCommand(loginCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

func checkSessionKey(sessionKey string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", DefaultServerUrl+"/api/v1/user/me", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Session", sessionKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("Failed to close response body: " + err.Error())
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("server responded with status code: %d", resp.StatusCode)
	}

	return nil
}

func handleURLShortening(urlToShorten string) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get user home directory: " + err.Error())
	}

	sessionFilePath := filepath.Join(homeDir, ".goshort", "session.txt")
	sessionKey, err := os.ReadFile(sessionFilePath)
	if err != nil {
		log.Fatal("Failed to read session file: " + err.Error())
	}

	log.Println("Checking session token...")

	if err := checkSessionKey(string(sessionKey)); err != nil {
		log.Fatal(err)
	}
	log.Println("Session token is valid!")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	shortenRequestBody := map[string]string{
		"url": urlToShorten,
	}

	shortenRequestBodyJson, err := json.Marshal(shortenRequestBody)
	if err != nil {
		log.Fatal("Failed to marshal shorten request body: " + err.Error())
	}

	req, err := http.NewRequest("POST", DefaultServerUrl+"/api/v1/shortlinks", bytes.NewBuffer(shortenRequestBodyJson))
	if err != nil {
		log.Fatal("Failed to create request: " + err.Error())
	}

	req.Header.Set("Session", string(sessionKey))
	req.Header.Set("Content-Type", "application/json")

	shortenResp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to shorten URL: " + err.Error())
	}

	if shortenResp.StatusCode != 200 {
		log.Fatal("Failed to shorten URL: " + shortenResp.Status)
	}

	type Shortlink struct {
		ID       int    `json:"id"`
		LongURL  string `json:"long_url"`
		ShortURL string `json:"short_url"`
	}

	type ShortenResponse struct {
		Message   string    `json:"message"`
		Shortlink Shortlink `json:"shortlink"`
		Success   bool      `json:"success"`
	}

	var shortenRespData ShortenResponse
	if err := json.NewDecoder(shortenResp.Body).Decode(&shortenRespData); err != nil {
		log.Fatal("Failed to decode shorten response body: " + err.Error())
	}

	if !shortenRespData.Success {
		log.Fatal("URL shortening failed: " + shortenRespData.Message)
	}

	log.Println("Shortened URL:", shortenRespData.Shortlink.ShortURL)
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal("Failed to close response body: " + err.Error())
		}
	}(shortenResp.Body)
}
