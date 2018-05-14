package spark

import (
	"encoding/json"
	"net/http"
)

type Client struct {
}

type Application struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Attempts []Attempt `json:"attempts"`
}

type Attempt struct {
	StartTime string `json:"startTime"`
	EndTime   string `json:"endTime"`
	Completed bool   `json:"completed"`
}

func (c *Client) GetApplicationLogs(masterAddress string) ([]Application, error) {
	resp, err := http.Get("http://" + masterAddress + ":4040/api/v1/applications")
	if err != nil {
		return nil, err
	}

	var apps []Application

	err = json.NewDecoder(resp.Body).Decode(&apps)
	if err != nil {
		return apps, err
	}

	return apps, nil
}
