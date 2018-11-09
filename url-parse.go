package main

import (
	"regexp"
)

// ExtractBBParts extract bitbucket link to get only the path up to the id
func ExtractBBParts(text string) []string {
	re := regexp.MustCompile(`bitbucket[a-z/.]*(/zume/[a-z]*/)pull-requests/(\d+)`)
	return re.FindStringSubmatch(text)
}

// ExtractMentionMe finds @tran.tu
func ExtractMentionMe(text string) []string {
	re := regexp.MustCompile(`@tran.tu`)
	return re.FindStringSubmatch(text)
}
