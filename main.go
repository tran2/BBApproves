package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
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
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(body))
	resBody := &zapierSlackMessage{}
	err = json.Unmarshal(body, resBody)
	if err != nil {
		fmt.Println(err)
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
	response.Write([]byte("Ok"))
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
	for _, userEmail := range allowedUsers {
		fmt.Println("userEmail " + userEmail + " " + email)
		if userEmail == email {
			return true
		}
	}
	return false
}

func redirectPolicyFunc() {

}

func approvesPR(path string) {
	client := &http.Client{}

	bbUser := "***REMOVED***"
	bbAppPassword := "***REMOVED***"
	bbAPI := "https://api.bitbucket.org/2.0/repositories"

	pullRequestURL := bbAPI + path
	fmt.Println("pullRequestUrl", pullRequestURL)

	req, err := http.NewRequest("GET", pullRequestURL, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.SetBasicAuth(bbUser, bbAppPassword)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	response, err := ioutil.ReadAll(resp.Body)
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

	approveRequest, err := http.NewRequest("POST", pullRequest.Links.Approve.Href, nil)
	if err != nil {
		fmt.Println(err)
	}

	approveRequest.SetBasicAuth(bbUser, bbAppPassword)
	approveResponse, err := client.Do(approveRequest)
	if err != nil {
		fmt.Println(err)
	}

	if approveResponse.StatusCode == http.StatusOK {
		fmt.Println("Approved!")
	} else {
		fmt.Println("Failed")
		fmt.Println(approveResponse)
	}
}

func main() {
	fmt.Println("starting server at :8001")
	http.HandleFunc("/webhook", webhookHandler)
	http.ListenAndServe(":8001", nil)
}
