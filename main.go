package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

func main() {
	sitePtr := flag.String("site", "", "the site to check")
	flag.Parse()

	if *sitePtr == "" {
		fmt.Println("Please provide a site to check using the -site flag")
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", *sitePtr, nil)
	if err != nil {
		fmt.Println("Error while making request:", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error while making request:", err)
		return
	}
	defer resp.Body.Close()

	var state string
	bodyBytes := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			bodyBytes = append(bodyBytes, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	bodyString := string(bodyBytes)
	if isStateful(bodyString) {
		state = "Stateful"
	} else if isStateless(bodyString) {
		state = "Stateless"
	} else {
		state = "Not sure"
	}

	fmt.Printf("The website %s is %s", *sitePtr, state)
}

func isStateful(body string) bool {
	// Check for cookies
	if strings.Contains(body, "Set-Cookie") {
		return true
	}

	// Check for session ID in URL
	re := regexp.MustCompile(`([?&])[^=]+=[^&]*PHPSESSID=[^&]+`)
	if re.MatchString(body) {
		return true
	}

	// Check for hidden form fields
	re = regexp.MustCompile(`type=["']hidden["']`)
	if re.MatchString(body) {
		return true
	}

	// Check for JavaScript code that sets cookies or modifies form data
	re = regexp.MustCompile(`(?s)<script.*?</script>`)
	scripts := re.FindAllString(body, -1)
	for _, script := range scripts {
		if strings.Contains(script, "document.cookie") || strings.Contains(script, ".value") {
			return true
		}
	}

	// Check for HTTP methods other than GET and POST
	re = regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
	if re.MatchString(body) {
		return true
	}

	// Check for AJAX requests
	re = regexp.MustCompile(`(?s)<script.*?>.*?xmlhttprequest.*?</script>`)
	if re.MatchString(body) {
		return true
	}

	// Check for WebSocket connections
	re = regexp.MustCompile(`(?s)<script.*?>.*?websocket.*?</script>`)
	if re.MatchString(body) {
		return true
	}

	return false
}

func isStateless(body string) bool {
	// Check for RESTful URLs
	re := regexp.MustCompile(`\/[a-zA-Z0-9_-]+`)
	if re.MatchString(body) {
		return true
	}

	// Check for URL query parameters
	if strings.Contains(body, "?") {
		return true
	}

	// Check for HTTP methods other than GET and POST
	re = regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
	if re.MatchString(body) {
		return true
	}

	// Check for AJAX requests
	return false
}
