package problem8

import "math/bits"

type Cypher interface {
	operation(input byte, pos int) byte
}

type ReverseCypher struct{}

type XORNCypher struct {
	value byte
}

type XORPosCypher struct{}

type AddNCypher struct {
	value byte
}

type AddPosCypher struct{}

func (c ReverseCypher) operation(input byte, pos int) byte {
	return byte(bits.Reverse(uint(input)))
}

func (c XORNCypher) operation(input byte, pos int) byte {
	return input ^ c.value
}

func (c XORPosCypher) operation(input byte, pos int) byte {
	return input ^ byte(pos)
}

func (c AddNCypher) operation(input byte, pos int) byte {
	a, _ := bits.Add(uint(input), uint(c.value), 0)
	return byte(a)
}

func (c AddPosCypher) operation(input byte, pos int) byte {
	a, _ := bits.Add(uint(input), uint(pos), 0)
	return byte(a)
}
