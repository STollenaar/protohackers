package main

import (
	"os"
	problem0 "protohackers/lib/problem-0"
	problem1 "protohackers/lib/problem-1"
)

var problemSelection string

func init() {
	problemSelection = os.Getenv("PROBLEM")
	if problemSelection == "" {
		problemSelection = "problem-0"
	}
}

func main() {
	switch problemSelection {
	case "problem-1":
		problem1.Problem()
	default:
		problem0.Problem()
	}
}
