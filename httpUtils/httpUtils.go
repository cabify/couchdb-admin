package httpUtils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/apex/log"
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

func (a *AuthenticatedHttpRequester) RunRequest(req *http.Request, dest interface{}) error {
	req.SetBasicAuth(a.username, a.password)

	log.WithFields(log.Fields{"URL": req.URL, "method": req.Method}).Debug("Sending request...")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("Received response %d for %s", resp.StatusCode, req.URL.String())
	}

	defer resp.Body.Close()

	if dest != nil {
		if err = json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return err
		}
	}

	return nil
}

func (a *AuthenticatedHttpRequester) Server() string {
	return a.server
}
