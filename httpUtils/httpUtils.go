package httpUtils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type AuthenticatedHttpRequester struct {
	server, username, password string
	httpClient                 *http.Client
}

func NewAuthenticatedHttpRequester(username, password, server string) (ahr *AuthenticatedHttpRequester) {
	ahr = &AuthenticatedHttpRequester{
		server:   server,
		username: username,
		password: password,
		httpClient: &http.Client{
			Timeout: time.Second * 10,
		},
	}
	return
}

func (a *AuthenticatedHttpRequester) RunRequest(req *http.Request, dest interface{}) {
	req.SetBasicAuth(a.username, a.password)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode/100 != 2 {
		log.Fatalf("Received response %d for %s", resp.StatusCode, req.URL.String())
	}

	defer resp.Body.Close()

	if dest != nil {
		if err = json.NewDecoder(resp.Body).Decode(dest); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Response empty: %v\n", resp)
	}
}

func (a *AuthenticatedHttpRequester) GetServer() string {
	return a.server
}
