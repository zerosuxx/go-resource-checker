package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/zerosuxx/go-resource-checker/pkg/checker"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

var Version = "development"

func main() {
	log.Println("Res0urce Checker " + Version)

	serverCommand := flag.NewFlagSet("server", flag.ContinueOnError)
	var address string
	serverCommand.StringVar(&address, "addr", "0.0.0.0:8000", "Server address")

	checkCommand := flag.NewFlagSet("check", flag.ContinueOnError)
	var resourceUrl string
	var timeout int
	checkCommand.StringVar(&resourceUrl, "url", "", "Resource url (eg: tcp://localhost:1234)")
	checkCommand.IntVar(&timeout, "timeout", 30, "Timeout")

	if len(os.Args) < 2 {
		serverCommand.Usage()
		checkCommand.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		if err := checkCommand.Parse(os.Args[2:]); err == nil {
			if resourceUrl == "" {
				checkCommand.Usage()
				os.Exit(1)
			}
			handleCheckCommand(resourceUrl, timeout)
		}
	default:
		if err := serverCommand.Parse(os.Args[2:]); err == nil {
			if address == "" {
				serverCommand.Usage()
				os.Exit(1)
			}
			handleServerCommand(address, timeout)
		}
	}
}

func handleServerCommand(address string, timeout int) {
	var resourceUrls []string
	_ = json.Unmarshal([]byte(os.Getenv("RESOURCE_URLS")), &resourceUrls)
	slackNotificationUrl := os.Getenv("SLACK_WEBHOOK_URL")
	successValue := os.Getenv("FORCE_SUCCESS_RESPONSE") == "1"
	resourceChecker := checker.ResourceChecker{CheckSuccessOnHealthCheck: true}

	log.Println(resourceUrls)

	http.HandleFunc("/check", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Auth-Token") != os.Getenv("AUTH_TOKEN") {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		checkUrl := request.URL.Query().Get("url")
		u, err := url.Parse(checkUrl)
		if checkUrl == "" || err != nil {
			writer.WriteHeader(http.StatusBadRequest)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)

		log.Println("Checking: " + u.String() + " (max timeout: " + strconv.Itoa(timeout) + ")")

		startTime := time.Now()
		connectionError := resourceChecker.Check(u, timeout)
		checkDuration := time.Now().Sub(startTime)

		log.Println("Checked: " + u.String() + " (duration: " + checkDuration.String() + ")")

		response := checker.JsonResponse{}
		if connectionError != nil {
			response.Success = false
			response.Message = connectionError.Error()
		} else {
			response.Success = true
		}

		responseByte, _ := json.Marshal(response)
		_, _ = writer.Write(responseByte)
	})

	http.HandleFunc("/healthcheck", func(writer http.ResponseWriter, request *http.Request) {
		if request.Header.Get("X-Auth-Token") != os.Getenv("AUTH_TOKEN") {
			writer.WriteHeader(http.StatusUnauthorized)

			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)

		response := checker.JsonResponse{}
		success := true
		for _, resourceUrl := range resourceUrls {
			u, _ := url.Parse(resourceUrl)

			log.Println("Checking: "+resourceUrl+" (max timeout:", strconv.Itoa(timeout)+")")

			connectionError := resourceChecker.Check(u, timeout)

			if connectionError != nil {
				success = successValue
				if slackNotificationUrl != "" {
					slackResponse := sendSlackNotification(slackNotificationUrl, "*"+resourceUrl+" is not healthy!*")
					log.Println("Slack response: " + slackResponse)
				}
				log.Println("Error: " + connectionError.Error())
			} else {
				log.Println("Ok: " + resourceUrl)
			}
		}

		response.Success = success
		responseByte, _ := json.Marshal(response)
		_, _ = writer.Write(responseByte)
	})

	log.Println("Server listening on: http://" + address)
	log.Fatal(http.ListenAndServe(address, nil))
}

func handleCheckCommand(resourceUrl string, timeout int) {
	u, err := url.Parse(resourceUrl)
	if err != nil {
		panic(err)
	}

	log.Println("Connecting to:", u.Scheme+"://"+u.Host+" (timeout:", strconv.Itoa(timeout)+")")

	resourceChecker := checker.ResourceChecker{}
	connectionError := resourceChecker.Check(u, timeout)

	if connectionError != nil {
		log.Println(connectionError)
		os.Exit(2)
	}

	log.Println("Connection successfully")
}

func sendSlackNotification(url string, message string) string {
	client := http.Client{}
	values := map[string]string{"text": message}
	jsonValue, _ := json.Marshal(values)

	response, _ := client.Post(url, "Content-type: application/json", bytes.NewBuffer(jsonValue))
	body, _ := io.ReadAll(response.Body)

	return string(body)
}
