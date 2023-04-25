package main

import (
	"flag"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"go.uber.org/zap"
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

	fmt.Printf("\nThe website %s is %s", *sitePtr, state)
}

var logger = zap.NewExample()
var log = logger.Sugar()

func isStateful(body string) bool {
	var wg sync.WaitGroup
	var hasCookies, hasSessionID, hasHiddenFields, hasJsCookiesOrFormData, hasNonStandardHttpMethods, hasAjaxRequests, hasWebSockets bool

	// Check for cookies
	wg.Add(1)
	go func() {
		defer wg.Done()
		if strings.Contains(body, "cookie") {
			hasCookies = true
		}
		log.Infoln("Set Cookies", hasCookies)
	}()

	// Check for session ID in URL
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`([?&])[^=]+=[^&]*PHPSESSID=[^&]+`) //TODO
		if re.MatchString(body) {
			hasSessionID = true
		}
		log.Infoln("Session ID", hasSessionID)
	}()

	// Check for hidden form fields
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`type=["']hidden["']`)
		if re.MatchString(body) {
			hasHiddenFields = true
		}
		log.Infoln("Hidden Fields", hasHiddenFields)
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
		log.Infoln("JS Cookies or Form Data", hasJsCookiesOrFormData)
	}()

	// Check for HTTP methods other than GET and POST
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
		if re.MatchString(body) {
			hasNonStandardHttpMethods = true
		}
		log.Infoln("Non Standard HTTP Methods", hasNonStandardHttpMethods)
	}()

	// Check for AJAX requests
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`(?s)<script.*?>.*?xmlhttprequest.*?</script>`)
		if re.MatchString(body) {
			hasAjaxRequests = true
		}
		log.Infoln("AJAX Requests", hasAjaxRequests)
	}()

	// Check for WebSocket connections
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`(?s)<script.*?>.*?websocket.*?</script>`)
		if re.MatchString(body) {
			hasWebSockets = true
		}
		log.Infoln("Web Sockets", hasWebSockets)
	}()

	// Wait for all the goroutines to finish
	wg.Wait()

	count := 0
	if hasSessionID {
		count++
	}
	if hasHiddenFields {
		count++
	}
	if hasJsCookiesOrFormData {
		count++
	}
	if hasNonStandardHttpMethods {
		count++
	}
	if hasAjaxRequests {
		count++
	}
	if hasWebSockets {
		count++
	}
	if hasCookies {
		count++
	}

	println("Stateful checks: ", count)
	// If more than 2 of the above are true, then the site is stateful
	return count >= 2
}

func isStateless(body string) bool {
	var wg sync.WaitGroup
	var hasRestfulUrls, hasUrlQueryParams, hasNonStandardHttpMethods bool

	// Check for RESTful URLs
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`\/[a-zA-Z0-9_-]+`)
		if re.MatchString(body) {
			hasRestfulUrls = true
		}
		log.Infoln("RESTful URLs", hasRestfulUrls)
	}()

	// Check for URL query parameters
	wg.Add(1)
	go func() {
		defer wg.Done()
		if strings.Contains(body, "?") {
			hasUrlQueryParams = true
		}
		log.Infoln("URL Query Params", hasUrlQueryParams)
	}()

	// Check for HTTP methods other than GET and POST
	wg.Add(1)
	go func() {
		defer wg.Done()
		re := regexp.MustCompile(`method=["'](PUT|DELETE|PATCH|OPTIONS|HEAD|TRACE|CONNECT)["']`)
		if re.MatchString(body) {
			hasNonStandardHttpMethods = true
		}
		log.Infoln("Non Standard HTTP Methods", hasNonStandardHttpMethods)
	}()

	// Wait for all the goroutines to finish
	wg.Wait()

	count := 0
	if hasRestfulUrls {
		count++
	}
	if hasUrlQueryParams {
		count++
	}
	if hasNonStandardHttpMethods {
		count++
	}

	println("Stateless checks: ", count)
	// If more than 1 of the above are true, then the site is stateless
	return count >= 1
}
