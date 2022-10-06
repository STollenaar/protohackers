package main

import (
	"log"
	"os"
	problem0 "protohackers/lib/problem-0"
	problem1 "protohackers/lib/problem-1"
	problem2 "protohackers/lib/problem-2"
	problem3 "protohackers/lib/problem-3"
	problem4 "protohackers/lib/problem-4"
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
	case "problem-0":
		problem0.Problem()
	case "problem-1":
		problem1.Problem()
	case "problem-2":
		problem2.Problem()
	case "problem-3":
		problem3.Problem()
	case "problem-4":
		problem4.Problem()
	default:
		log.Fatal("Problem not found")
	}
}
