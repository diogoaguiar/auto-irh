package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	baseURL = "https://irh.pt/irh/"
)

var (
	user    string
	pass    string
	client  = &http.Client{}
	session string
	expires time.Time
)

func main() {
	log.Println("Starting...")
	log.Println("Loading config...")
	config, err := os.ReadFile("config.ini")
	if err != nil {
		log.Fatal(err)
	}

	// Parse user and pass from config file
	_, err = fmt.Sscanf(string(config), "%s %s", &user, &pass)
	if err != nil {
		log.Fatal(err)
	}

	punch()

	log.Println("Done")
}

func login() error {
	// Check if session is still valid
	if loadData() {
		// Session is still valid
		log.Println("Session is still valid")
		return nil
	}

	// Session is not valid
	log.Println("Session is not valid")

	req, err := http.NewRequest("POST", baseURL+"/login/login_c/login", nil)
	if err != nil {
		return err
	}
	req.PostForm = url.Values{
		"user":  {user},
		"senha": {pass},
	}

	log.Println("Logging in...")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Get the cookie named "ci_session"
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "ci_session" {
			session = cookie.Value
			expires = cookie.Expires
			log.Println("Logged in")
			break
		}
	}

	// Save session and expires to file called "data"
	log.Println("Saving session data...")
	data := fmt.Sprintf("%s %s", session, expires.Format(time.RFC3339))
	err = os.WriteFile("data", []byte(data), 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadData() bool {
	// Read session and expires from file called "data"
	data, err := os.ReadFile("data")
	if err != nil {
		log.Println(err)
	}

	// Check if file is empty
	if len(data) == 0 {
		log.Println("File is empty or does not exist")
		return false
	}

	// Parse session and expires
	var expiresString string
	_, err = fmt.Sscanf(string(data), "%s %s", &session, &expiresString)
	if err != nil {
		log.Println(err)
		return false
	}
	expires, err = time.Parse(time.RFC3339, expiresString)
	if err != nil {
		log.Println(err)
		return false
	}

	return time.Now().Before(expires)
}

func isLoggedIn() bool {
	return session != ""
}

func punch() bool {
	if !isLoggedIn() {
		err := login()
		if err != nil {
			log.Println(err)
			return false
		}
	}

	req, err := http.NewRequest("POST", baseURL+"/picagens/picagens_c/registarPicagem", nil)
	if err != nil {
		log.Println(err)
		return false
	}
	req.Header.Set("Cookie", "ci_session="+session)

	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return false
	}

	if resp.StatusCode != 200 || len(body) <= 3 {
		log.Println("Failed to punch")
		return false
	}

	log.Println("Punched successfully")

	return true
}
