package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func PostRequest(url string, apiKey string, data interface{}) (int, error) {

	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling system info to JSON:", err.Error())
		return 500, err
	}
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Println("Error creating request:", err.Error())
		return 500, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err.Error())
		return 500, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("%s", body)
	}
	return resp.StatusCode, nil
}

func PutRequest(url string, apiKey string, data interface{}) (int, error) {

	client := &http.Client{}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error marshalling system info to JSON:", err.Error())
		return 500, err
	}
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(jsonData)))
	if err != nil {
		log.Println("Error creating request:", err.Error())
		return 500, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err.Error())
		return 500, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("%s", body)
	}
	return resp.StatusCode, nil
}

func GetRequest(url string, apiKey string) (int, string, error) {
	// Send a GET request to the specified URL with the API key
	// Returns the status code and response body as a string

	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println("Error creating request:", err.Error())
		return 500, "", err
	}
	req.Header.Set("x-api-key", apiKey)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err.Error())
		return 500, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return resp.StatusCode, "", nil
	}
	body, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(body), nil
}
