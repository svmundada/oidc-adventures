package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	oidcTokenPathEnvVar string = "OIDC_TOKEN_PATH"
	serverURLEnvVar     string = "OIDC_REQUIRED_SERVER"
)

func getToken(tokenPath string) (string, error) {
	tokenBytes, err := os.ReadFile(tokenPath)
	if err != nil {
		return "", fmt.Errorf("unable to extract oidc token, %v", err)
	}
	return string(tokenBytes), nil
}

func main() {
	tokenPath := os.Getenv(oidcTokenPathEnvVar)
	serverUrl := os.Getenv(serverURLEnvVar)
	// every 10 seconds make a request to server
	echoPath := strings.Join([]string{serverUrl, "echo"}, "/")
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		echoPath, nil)
	if err != nil {
		fmt.Printf("request construction failed: %v", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(time.Second * 10)
	done := make(chan struct{})
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			fmt.Println(t)
			token, err := getToken(tokenPath)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			req.Header.Add("Authorization", token)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Printf("request failed: %v", err)
				os.Exit(1)
			}
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("statusCode: %d, body: %s\n", resp.StatusCode, string(body))
			defer resp.Body.Close()
		}
	}

}
