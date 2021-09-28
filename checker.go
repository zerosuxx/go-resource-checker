package main

import (
	"encoding/json"
	"flag"
	"github.com/zerosuxx/go-resource-checker/pkg/checker"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
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
	resourceChecker := checker.ResourceChecker{CheckSuccessOnHealthCheck: true}

	log.Println(resourceUrls)

	http.HandleFunc("/healthcheck", func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(200)

		response := checker.SuccessResponse{}
		success := true
		for _, resourceUrl := range resourceUrls {
			u, _ := url.Parse(resourceUrl)

			log.Println("Checking: "+resourceUrl+" (timeout:", strconv.Itoa(timeout)+")")

			connectionError := resourceChecker.Check(u, timeout)

			if connectionError != nil {
				success = false
				log.Println("Error: " + connectionError.Error())
			} else {
				log.Println("Ok: " + resourceUrl)
			}
		}

		response.Success = success
		responseByte, _ := json.Marshal(response)
		_, _ = writer.Write(responseByte)
	})

	log.Println("Server listening on: " + address)
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
