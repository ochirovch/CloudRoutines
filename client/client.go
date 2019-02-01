package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Bucket contains list of URLs and parameters for handling
type Bucket struct {
	ID    string
	Paths []string
}

// GetListURLs get list addresses and visit their
func GetListURLs(address string) (bucket Bucket, err error) {
	resp, err := http.Get(address + "/channel/send")
	if err != nil {
		return bucket, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &bucket)
	if err != nil {
		return bucket, err
	}
	return bucket, nil
}

func SendListURLs(ListURLs []string, url string) {
	jList, err := json.Marshal(ListURLs)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewReader(jList))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
