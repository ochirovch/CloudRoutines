package client

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Bucket contains list of URLs and parameters for handling
type Bucket struct {
	ID    string
	Paths []string
}

// GetTask get list addresses and visit their
func GetTask() (bucket <-chan Bucket, err error) {
	resp, err := http.Get(":8099/channel/gettask")
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

func SendResults() {

}
