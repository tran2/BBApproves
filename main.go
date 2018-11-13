package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type slackProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

type slackUser struct {
	Profile slackProfile `json:"profile"`
	Name    string       `json:"name"`
}

type zapierSlackMessage struct {
	Text string    `json:"text"`
	User slackUser `json:"user"`
}

type href struct {
	Href string `json:"href"`
}

type bBLinks struct {
	Diff    href `json:"diff"`
	Approve href `json:"approve"`
}

type pullRequestResponse struct {
	Links bBLinks `json:"links"`
}

func webhookHandler(response http.ResponseWriter, request *http.Request) {
	fmt.Println(request)
	if request.Header.Get("Authorization") != "***REMOVED***" {
		response.WriteHeader(http.StatusUnauthorized)
		response.Write([]byte("Unauthorized"))
		return
	}

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
		http.Error(response, "Bad Request", http.StatusBadRequest)
		return
	}
	fmt.Println("request body: " + string(body))
	resBody := &zapierSlackMessage{}
	err = json.Unmarshal(body, resBody)
	if err != nil {
		fmt.Println(err)
		http.Error(response, "Bad Request", http.StatusBadRequest)
		return
	}

	fmt.Println(resBody)
	fmt.Println(resBody.User.Profile.Email)

	path, err := getBBPath(resBody)
	if err != nil {
		fmt.Println(err)
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte(err.Error()))
		return
	}

	approvesPR(path)
	response.WriteHeader(http.StatusOK)
	response.Write([]byte("Success"))
}

func getBBPath(slackMessage *zapierSlackMessage) (string, error) {
	text := slackMessage.Text
	if !isUserAllowed(slackMessage.User.Profile.Email) {
		return "", errors.New("Unsupported user")
	}

	// ["bitbucket.org/zume/admin/pull-requests/446" "/zume/admin/" "446"]
	bbParts := ExtractBBParts(text)
	mentionMe := ExtractMentionMe(text)

	if len(mentionMe) < 1 {
		return "", errors.New("Must mention Tran atleast 1 time")
	}

	if len(bbParts) < 3 {
		return "", errors.New("Invalid link must be bitbucket pull request link")
	}

	return bbParts[1] + "pullrequests/" + bbParts[2], nil
}

func isUserAllowed(email string) bool {
	allowedUsers := []string{"***REMOVED***", "***REMOVED***"}
	if len(os.Getenv("ALLOWED_USERS")) > 0 {
		allowedUsers = strings.Split(os.Getenv("ALLOWED_USERS"), ",")
	}
	str := fmt.Sprintf("allowedUsers %v", allowedUsers)
	fmt.Println(str)
	for _, userEmail := range allowedUsers {
		if userEmail == email {
			return true
		}
	}
	return false
}

func redirectPolicyFunc() {

}

func sendRequestToBitBucket(method string, url string) ([]byte, error) {
	bbUser := "***REMOVED***"
	bbAppPassword := "***REMOVED***"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	req.SetBasicAuth(bbUser, bbAppPassword)
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return []byte{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []byte("Was not ok"), err
	}

	response, err := ioutil.ReadAll(resp.Body)
	return response, err
}

func approvesPR(path string) {
	bbAPI := "https://api.bitbucket.org/2.0/repositories"

	pullRequestURL := bbAPI + path
	fmt.Println("pullRequestUrl", pullRequestURL)
	response, err := sendRequestToBitBucket("GET", pullRequestURL)
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(string(response))

	pullRequest := &pullRequestResponse{}
	err = json.Unmarshal(response, pullRequest)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(pullRequest.Links.Approve.Href)

	approveResponse, err := sendRequestToBitBucket("POST", pullRequest.Links.Approve.Href)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Failed!")
	}

	fmt.Println("Approved!")
	fmt.Println(approveResponse)
}

func main() {
	fmt.Println("starting server at :8001")
	http.HandleFunc("/webhook", webhookHandler)
	http.ListenAndServe(":8001", nil)
}
