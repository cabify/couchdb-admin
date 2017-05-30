package http_utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

var HttpClient *http.Client
var Username, Password string

func RunRequest(req *http.Request, dest interface{}) {
	req.SetBasicAuth(Username, Password)

	resp, err := HttpClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if resp.StatusCode/100 != 2 {
		log.Printf("Received response %d for %s", resp.StatusCode, req.URL.String())
	}

	defer resp.Body.Close()

	if dest != nil {
		if err = json.NewDecoder(resp.Body).Decode(dest); err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Println(resp)
	}
}
