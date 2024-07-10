package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/elazarl/goproxy"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port         string `yaml:"port"`
	TargetServer string `yaml:"target_server"`
}

func main() {
	// Load configuration file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Open or create log files
	requestLogFile, err := os.OpenFile("logs/request.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open request log file: %v", err)
	}
	defer requestLogFile.Close()

	responseLogFile, err := os.OpenFile("logs/response.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open response log file: %v", err)
	}
	defer responseLogFile.Close()

	errorLogFile, err := os.OpenFile("logs/error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open error log file: %v", err)
	}
	defer errorLogFile.Close()

	// Set log output destinations
	requestLogger := log.New(io.MultiWriter(os.Stdout, requestLogFile), "REQUEST: ", log.Ldate|log.Ltime|log.Lshortfile)
	responseLogger := log.New(io.MultiWriter(os.Stdout, responseLogFile), "RESPONSE: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(io.MultiWriter(os.Stderr, errorLogFile), "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Create goproxy instance
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true

	// Redirect proxy requests to custom target
	proxy.NonproxyHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		redirectHandler(w, req, config.TargetServer, requestLogger, responseLogger, errorLogger)
	})

	log.Printf("Starting proxy server on localhost:%s", config.Port)
	if err := http.ListenAndServe("localhost:"+config.Port, proxy); err != nil {
		errorLogger.Fatalf("ListenAndServe: %v", err)
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func redirectHandler(w http.ResponseWriter, r *http.Request, targetServer string, requestLogger, responseLogger, errorLogger *log.Logger) {
	// Log the request
	logRequest(r, requestLogger)

	// Parse the target server URL
	target, err := url.Parse(targetServer)
	if err != nil {
		http.Error(w, "Invalid target server URL", http.StatusInternalServerError)
		errorLogger.Printf("Invalid target server URL: %v", err)
		return
	}

	// Construct a new request URL by combining the target server and the original request path
	proxyURL := target.ResolveReference(r.URL)
	proxyReq, err := http.NewRequest(r.Method, proxyURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		errorLogger.Printf("Failed to create request: %v", err)
		return
	}

	// Copy and modify headers
	for name, values := range r.Header {
		for _, value := range values {
			// Example modification: add or modify a header
			if name == "User-Agent" {
				proxyReq.Header.Set(name, "MyCustomUserAgent")
			} else {
				proxyReq.Header.Add(name, value)
			}
		}
	}
	// Add a new header
	proxyReq.Header.Set("X-Added-Header", "HeaderValue")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		http.Error(w, "Failed to connect to server", http.StatusInternalServerError)
		errorLogger.Printf("Failed to connect to server: %v", err)
		return
	}
	defer resp.Body.Close()

	// Log the response
	logResponse(resp, responseLogger)

	// Forward the response to the client
	for key, value := range resp.Header {
		for _, v := range value {
			w.Header().Add(key, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		errorLogger.Printf("Failed to copy response body: %v", err)
	}
}

func logRequest(r *http.Request, logger *log.Logger) {
	logger.Printf("Request: %s %s", r.Method, r.URL)
	for name, values := range r.Header {
		for _, value := range values {
			logger.Printf("Request Header: %s: %s", name, value)
		}
	}

	// Log the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logger.Printf("Failed to read request body: %v", err)
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	logger.Printf("Request Body: %s", body)
}

func logResponse(resp *http.Response, logger *log.Logger) {
	logger.Printf("Response: %s", resp.Status)
	for name, values := range resp.Header {
		for _, value := range values {
			logger.Printf("Response Header: %s: %s", name, value)
		}
	}

	// Log the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Printf("Failed to read response body: %v", err)
		return
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	logger.Printf("Response Body: %s", body)
}
