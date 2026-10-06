package main

import (
	"metrics-study-project-go/internal/agent"
)

func main() {
	agent := agent.CreateAgent()

	agent.Run()
}
