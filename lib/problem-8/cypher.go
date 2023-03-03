package problem8

import (
	"fmt"
	"math/bits"
)

type Cypher interface {
	operation(input byte, encoding bool) byte
	reset()
	fmt.Stringer
}

type ReverseCypher struct{}

type XORNCypher struct {
	value byte
}

type XORPosCypher struct {
	pos byte
}

type AddNCypher struct {
	value byte
}

type AddPosCypher struct {
	pos byte
}

func (c *ReverseCypher) operation(input byte, encoding bool) byte {
	return bits.Reverse8(input)
}

func (c *ReverseCypher) reset() {}

func (c ReverseCypher) String() string {
	return "Reverse"
}

func (c *XORNCypher) operation(input byte, encoding bool) byte {
	return input ^ c.value
}

func (c *XORNCypher) reset() {}

func (c *XORNCypher) String() string {
	return fmt.Sprintf("XOR(%d)", c.value)
}

func (c *XORPosCypher) operation(input byte, encoding bool) byte {
	r := input ^ c.pos
	c.pos++
	return r
}

func (c *XORPosCypher) reset() {
	c.pos = 0
}

func (c *XORPosCypher) String() string {
	return fmt.Sprintf("XORPos{p:%d}", c.pos)
}

func (c *AddNCypher) operation(input byte, encoding bool) byte {
	if encoding {
		return input + c.value
	}
	return input - c.value
}

func (c *AddNCypher) reset() {}

func (c *AddNCypher) String() string {
	return fmt.Sprintf("ADD(%d)", c.value)
}

func (c *AddPosCypher) operation(input byte, encoding bool) byte {
	var r byte
	if encoding {
		r = input + c.pos
	} else {
		r = input - c.pos
	}
	c.pos++
	return r
}

func (c *AddPosCypher) reset() {
	c.pos = 0
}

func (c *AddPosCypher) String() string {
	return fmt.Sprintf("AddPos{p:%d}", c.pos)
}
