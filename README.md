## Golang Web State Detector
A simple tool written in Golang to detect the state (stateful or stateless) of a website. The tool analyzes various aspects of a website, such as cookies, caching, redirects, HTTP methods, AJAX requests, and WebSocket connections, to determine its state.

The tool uses regular expressions to match patterns in the website's response headers and body. It returns the state of the website as "stateful" or "stateless", along with the reason for its state. If the tool is unable to determine the website's state, it returns "not sure".

This tool is useful for analyzing websites and understanding their state, which can be helpful in various scenarios, such as security testing, performance testing, and debugging.

## Usage
To use the tool, simply provide the URL of the website you want to analyze. The tool will send an HTTP GET request to the website, analyze the response, and return the state of the website.

```
gor main.go -site https://vatsalchauhan.me/
```

## Contribution
Contributions are welcome and encouraged! If you find a bug, have a suggestion for improvement, or want to add a new feature, feel free to create a pull request or open an issue.

## License
This tool is released under the MIT License. See the LICENSE file for details.
