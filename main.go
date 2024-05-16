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

	"github.com/hashicorp/logutils"
)

type GetJobsRequest struct {
	URL   string `json:"url"`
	Token string `json:"auth"`
}

type GetAgentsRequest struct {
	URL   string `json:"url"`
	Token string `json:"auth"`
}

type GetAgentsResponse struct {
	Count  int     `json:"count"`
	Agents []Agent `json:"value"`
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

var (
	// url = "https://dev.azure.com/<project_name>/_apis/distributedtask/pools/<pool_id>/jobrequests"
	url string
	// token = "<personal_access_token>"
	token string
	// https://dev.azure.com/<project_name>/_apis/distributedtask/pools?api-version=7.2-preview.1
	// url_agent string
	agent  bool
	delete bool
)

func init() {
	flag.StringVar(&url, "url", "##", "URL to get agent jobs from")
	flag.StringVar(&token, "token", "##", "Token to authenticate with Azure DevOps")
	flag.BoolVar(&agent, "agent", false, "Return agents instead of jobs")
	flag.BoolVar(&delete, "delete", false, "Delete agents instead of getting them")

	// flag.StringVar(&url_agent, "url_agent", "##", "URL to get agents")

	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("DEBUG"),
		Writer:   os.Stderr,
	}
	log.SetFlags(log.Lshortfile)
	log.SetOutput(filter)
}

func getAgents(options GetAgentsRequest) (GetAgentsResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", options.URL+"agents", nil)
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

	// _, err = io.ReadAll(resp.Body)
	// body, err := os.ReadFile("resp.json")
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(string(body))
	agents := Agents{}
	err = json.Unmarshal(body, &agents)
	if err != nil {
		log.Fatalln(err)
	}

	return GetAgentsResponse{
		Count:  agents.Count,
		Agents: agents.Agents,
	}, nil
}

func getJobs(options GetJobsRequest) (GetJobsResponse, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", options.URL+"jobrequests", nil)
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
	// log.Println(resp.StatusCode)
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

	if agent {
		respAgents, err := getAgents(GetAgentsRequest{
			URL:   url,
			Token: token,
		})
		if err != nil {
			log.Fatalln(err)
		}
		m, _ := json.MarshalIndent(respAgents, "", "  ")
		fmt.Println(string(m))
		if delete {
			for _, agent := range respAgents.Agents {
				client := &http.Client{}
				req, err := http.NewRequest("DELETE", url+"agents/"+fmt.Sprint(agent.ID)+"?api-version=7.1-preview.1", nil)
				if err != nil {
					log.Fatalln(err)
				}

				// Add basic authentication header
				auth := ":" + token
				encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
				req.Header.Add("Authorization", "Basic "+encodedAuth)

				resp, err := client.Do(req)
				if err != nil {
					log.Fatalln(err)
				}
				defer resp.Body.Close()
				log.Println(resp.StatusCode)
				if resp.StatusCode == 204 {
					log.Println("Agent deleted successfully")
				} else {
					log.Println("Failed to delete agent")
				}

				// defer resp.Body.Close()
				// body, err := io.ReadAll(resp.Body)
				// if err != nil {
				// 	log.Fatalln(err)
				// }
				// log.Println(string(body))
			}
		}
	} else {
		respJobs, err := getJobs(GetJobsRequest{
			URL:   url,
			Token: token,
		})
		if err != nil {
			log.Fatalln(err)
		}
		m, _ := json.MarshalIndent(respJobs, "", "  ")
		fmt.Println(string(m))

	}
}
