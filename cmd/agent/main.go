package main

import (
	"flag"
	"metrics-study-project-go/internal/agent"
)

var flagRunAddr string
var flagSendInterval int
var flagCollectInterval int

func main() {
	flag.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&flagSendInterval, "r", 10, "metric send interval in seconds")
	flag.IntVar(&flagCollectInterval, "p", 2, "metric collect interval in seconds")

	flag.Parse()

	agent := agent.CreateAgent(flagRunAddr, flagSendInterval, flagCollectInterval)

	agent.Run()
}
