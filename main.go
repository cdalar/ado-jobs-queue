package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type GetJobsOptions struct {
	URL   string `json:"url"`
	Token string `json:"auth"`
}

type GetJobsResponse struct {
	Count struct {
		Total    int `json:"total"`
		Waiting  int `json:"waiting"`
		Running  int `json:"running"`
		Finished int `json:"finished"`
	} `json:"count"`
	Jobs []Job `json:"jobs"`
}

// func (g GetJobsResponse) String() string {
// 	return "Total: " + fmt.Sprint(g.Count.Total) + "\n" +
// 		"Waiting: " + fmt.Sprint(g.Count.Waiting) + "\n" +
// 		"Running: " + fmt.Sprint(g.Count.Running) + "\n" +
// 		"Finished: " + fmt.Sprint(g.Count.Finished) + "\n"
// }

var (
	url   string
	token string
)

func init() {
	flag.StringVar(&url, "url", "##", "URL to get agent jobs from")
	flag.StringVar(&token, "token", "##", "Token to authenticate with Azure DevOps")
}

func getJobs(options GetJobsOptions) (GetJobsResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", options.URL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	// Add basic authentication header
	auth := ":" + options.Token
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	// Don't forget to close the response body
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	// body, err := os.ReadFile("resp.json")
	if err != nil {
		log.Fatalln(err)
	}

	// Print the response body to stdout
	// fmt.Println(string(body))
	jobs := Jobs{}
	err = json.Unmarshal(body, &jobs)
	if err != nil {
		log.Fatalln(err)
	}
	getJobsResponse := GetJobsResponse{}
	for _, job := range jobs.Value {
		if job.ReceiveTime.IsZero() && job.FinishTime.IsZero() {
			getJobsResponse.Count.Waiting++
		} else if !job.ReceiveTime.IsZero() && job.FinishTime.IsZero() {
			getJobsResponse.Count.Running++
		} else if !job.ReceiveTime.IsZero() && !job.FinishTime.IsZero() {
			getJobsResponse.Count.Finished++
		}
	}
	getJobsResponse.Count.Total = len(jobs.Value)
	return GetJobsResponse{
		Count: getJobsResponse.Count,
		Jobs:  jobs.Value,
	}, nil
}

func main() {
	flag.Parse()
	if url == "##" || token == "##" {
		fmt.Println("Both url and token are required!")
		flag.PrintDefaults()
		os.Exit(1)
	}
	getJobsOptions := GetJobsOptions{
		URL:   url,
		Token: token,
	}
	resp, err := getJobs(getJobsOptions)
	if err != nil {
		log.Fatalln(err)
	}
	m, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(m))
}
