package request

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func MakeRequest(payload, url, method string) (string, error) {
	var reader io.Reader
	if payload != "" && payload != "{}" {
		reader = bytes.NewReader([]byte(payload))
	}

	request, err := http.NewRequest(method, url, reader)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	client := &http.Client{}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	return bodyString, nil
}
