package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
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
	var wg sync.WaitGroup
	var hasCookies, hasSessionID, hasHiddenFields, hasJsCookiesOrFormData, hasNonStandardHttpMethods, hasAjaxRequests, hasWebSockets bool

	// Check for cookies
	wg.Add(1)
	go func() {
		defer wg.Done()
		if strings.Contains(body, "Set-Cookie") {
			hasCookies = true
		}
	}()

	// Check for session ID in URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`([?&])[^=]+=[^&]*PHPSESSID=[^&]+`)
		if re.MatchString(body) {
			hasSessionID = true
		}
	}()

	// Check for hidden form fields
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`type=["']hidden["']`)
		if re.MatchString(body) {
			hasHiddenFields = true
		}
	}()

	// Check for JavaScript code that sets cookies or modifies form data
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`(?s)<script.*?</script>`)
		scripts := re.FindAllString(body, -1)
		for _, script := range scripts {
			if strings.Contains(script, "document.cookie") || strings.Contains(script, ".value") {
				hasJsCookiesOrFormData = true
				break
			}
		}
	}()

	// Check for HTTP methods other than GET and POST
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
		if re.MatchString(body) {
			hasNonStandardHttpMethods = true
		}
	}()

	// Check for AJAX requests
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`(?s)<script.*?>.*?xmlhttprequest.*?</script>`)
		if re.MatchString(body) {
			hasAjaxRequests = true
		}
	}()

	// Check for WebSocket connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`(?s)<script.*?>.*?websocket.*?</script>`)
		if re.MatchString(body) {
			hasWebSockets = true
		}
	}()

	// Wait for all the goroutines to finish
	wg.Wait()

	return hasCookies && hasSessionID && hasHiddenFields && hasJsCookiesOrFormData && hasNonStandardHttpMethods && hasAjaxRequests && hasWebSockets
}

func isStateless(body string) bool {
	var wg sync.WaitGroup
	var hasRestfulUrls, hasUrlQueryParams, hasNonStandardHttpMethods, hasAjaxRequests bool

	// Check for RESTful URLs
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`\/[a-zA-Z0-9_-]+`)
		if re.MatchString(body) {
			hasRestfulUrls = true
		}
	}()

	// Check for URL query parameters
	wg.Add(1)
	go func() {
		defer wg.Done()
		if strings.Contains(body, "?") {
			hasUrlQueryParams = true
		}
	}()

	// Check for HTTP methods other than GET and POST
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
		if re.MatchString(body) {
			hasNonStandardHttpMethods = true
		}
	}()

	// Wait for all the goroutines to finish
	wg.Wait()

	return hasRestfulUrls && hasUrlQueryParams && hasNonStandardHttpMethods && hasAjaxRequests
}
