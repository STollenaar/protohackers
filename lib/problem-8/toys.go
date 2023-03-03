package problem8

import (
	"fmt"
	"strconv"
	"strings"
)

type Toy struct {
	amount int
	kind   string
}

func createToys(line string) (toys []Toy) {
	for _, l := range strings.Split(line, ",") {
		args := strings.Split(l, " ")
		a := strings.ReplaceAll(args[0], "x", "")
		amount, _ := strconv.Atoi(a)
		t := Toy{
			amount: amount,
			kind:   strings.Join(args[1:], " "),
		}
		toys = append(toys, t)
	}
	return toys
}

func findMaxToy(toys []Toy) (toy Toy) {
	toy = toys[0]

	for _, t := range toys {
		if t.amount > toy.amount {
			toy = t
		}
	}
	return toy
}

func (t Toy) toString() string {
	return fmt.Sprintf("%dx %s\n", t.amount, t.kind)
}
