package main

import (
	"log"
	"os"
	problem0 "protohackers/lib/problem-0"
	problem1 "protohackers/lib/problem-1"
	problem10 "protohackers/lib/problem-10"
	problem2 "protohackers/lib/problem-2"
	problem3 "protohackers/lib/problem-3"
	problem4 "protohackers/lib/problem-4"
	problem5 "protohackers/lib/problem-5"
	problem6 "protohackers/lib/problem-6"
	problem7 "protohackers/lib/problem-7"
	problem8 "protohackers/lib/problem-8"
	problem9 "protohackers/lib/problem-9"
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
	case "problem-5":
		problem5.Problem()
	case "problem-6":
		problem6.Problem()
	case "problem-7":
		problem7.Problem()
	case "problem-8":
		problem8.Problem()
	case "problem-9":
		problem9.Problem()
	case "problem-10":
		problem10.Problem()
	default:
		log.Fatal("Problem not found")
	}
}
