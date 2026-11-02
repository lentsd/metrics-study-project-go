package main

import (
	"flag"
	"fmt"
	"log"
	"metrics-study-project-go/internal/agent"
	"os"
	"strconv"
)

var flagRunAddr string
var flagSendInterval int
var flagCollectInterval int

func main() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagSendInterval, "r", 10, "metric send interval in seconds")
	flag.IntVar(&flagCollectInterval, "p", 2, "metric collect interval in seconds")

	flag.Parse()

	setupEnv()

	fmt.Println(flagRunAddr, flagSendInterval, flagCollectInterval)

	agent := agent.CreateAgent(flagRunAddr, flagSendInterval, flagCollectInterval)

	agent.Run()
}

func setupEnv() {
	address, okAddress := os.LookupEnv("ADDRESS")
	sendInterval, okSend := os.LookupEnv("REPORT_INTERVAL")
	collectInterval, okCollect := os.LookupEnv("POLL_INTERVAL")

	if okAddress {
		flagRunAddr = address
	}

	if okSend {
		if send, err := strconv.Atoi(sendInterval); err == nil {
			flagSendInterval = send
		} else {
			log.Fatal(err)
		}
	}

	if okCollect {
		if collect, err := strconv.Atoi(collectInterval); err == nil {
			flagCollectInterval = collect
		} else {
			log.Fatal(err)
		}
	}
}
