package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var DefaultServerUrl = "https://shortdemo.tkbstudios.com"

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

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Failed to get user home directory: " + err.Error())
		}

		goshortDir := filepath.Join(homeDir, ".goshort")
		if err := os.MkdirAll(goshortDir, os.ModePerm); err != nil {
			log.Fatal("Failed to create .goshort directory: " + err.Error())
		}

		sessionFilePath := filepath.Join(goshortDir, "session.txt")
		sessionData := fmt.Sprintf("%s\n%s", serverUrl, loginRespData.Session)
		if err := os.WriteFile(sessionFilePath, []byte(sessionData), 0600); err != nil {
			log.Fatal("Failed to write session file: " + err.Error())
		}

		log.Println("Session data saved to:", sessionFilePath)

		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal("Failed to close response body: " + err.Error())
			}
		}(loginResp.Body)
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all shortlinks",
	Run: func(cmd *cobra.Command, args []string) {
		serverUrl, sessionKey, err := loadSessionData()
		if err != nil {
			log.Fatal(err)
		}

		color.Yellow("Checking session token...")

		if err := checkSessionKey(sessionKey, serverUrl); err != nil {
			log.Fatal(err)
		}
		color.Green("Session token is valid!")

		s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
		s.Start()

		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequest("GET", serverUrl+"/api/v1/shortlinks", nil)
		if err != nil {
			log.Fatal("Failed to create request: " + err.Error())
		}

		req.Header.Set("Session", sessionKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal("Failed to retrieve shortlinks: " + err.Error())
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Fatal("Failed to close response body: " + err.Error())
			}
		}(resp.Body)

		s.Stop()

		if resp.StatusCode != 200 {
			log.Fatal("Failed to retrieve shortlinks: " + resp.Status)
		}

		type Analytics struct {
			Browsers []struct {
				Browser string `json:"browser"`
				Count   int    `json:"count"`
				Country string `json:"country"`
			} `json:"browsers"`
			Clicks int `json:"clicks"`
		}

		type Shortlink struct {
			ID        int       `json:"id"`
			LongURL   string    `json:"long_url"`
			ShortURL  string    `json:"short_url"`
			Analytics Analytics `json:"analytics"`
		}

		type ListResponse struct {
			Message    string      `json:"message"`
			Shortlinks []Shortlink `json:"shortlinks"`
			Success    bool        `json:"success"`
		}

		var listRespData ListResponse
		if err := json.NewDecoder(resp.Body).Decode(&listRespData); err != nil {
			log.Fatal("Failed to decode list response body: " + err.Error())
		}

		if !listRespData.Success {
			log.Fatal("Failed to retrieve shortlinks: " + listRespData.Message)
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"ID", "Long URL", "Short URL", "Clicks", "Browser", "Count", "Country"})

		for _, link := range listRespData.Shortlinks {
			if len(link.Analytics.Browsers) > 0 {
				for i, browser := range link.Analytics.Browsers {
					if i == 0 {
						table.Append([]string{fmt.Sprintf("%d", link.ID), link.LongURL, link.ShortURL, fmt.Sprintf("%d", link.Analytics.Clicks), browser.Browser, fmt.Sprintf("%d", browser.Count), browser.Country})
					} else {
						table.Append([]string{"", "", "", "", browser.Browser, fmt.Sprintf("%d", browser.Count), browser.Country})
					}
				}
			} else {
				table.Append([]string{fmt.Sprintf("%d", link.ID), link.LongURL, link.ShortURL, fmt.Sprintf("%d", link.Analytics.Clicks), "", "", ""})
			}
		}

		table.Render()
	},
}

func main() {
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(listCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Command failed: %v", err)
	}
}

func loadSessionData() (string, string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	sessionFilePath := filepath.Join(homeDir, ".goshort", "session.txt")
	sessionData, err := os.ReadFile(sessionFilePath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read session file: %w", err)
	}

	lines := bytes.SplitN(sessionData, []byte("\n"), 2)
	if len(lines) < 2 {
		return "", "", fmt.Errorf("session file format is invalid")
	}

	serverUrl := string(lines[0])
	sessionKey := string(lines[1])
	return serverUrl, sessionKey, nil
}

func checkSessionKey(sessionKey string, serverUrl string) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", serverUrl+"/api/v1/user/me", nil)
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
	serverUrl, sessionKey, err := loadSessionData()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Checking session token...")

	if err := checkSessionKey(sessionKey, serverUrl); err != nil {
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

	req, err := http.NewRequest("POST", serverUrl+"/api/v1/shortlinks", bytes.NewBuffer(shortenRequestBodyJson))
	if err != nil {
		log.Fatal("Failed to create request: " + err.Error())
	}

	req.Header.Set("Session", sessionKey)
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
