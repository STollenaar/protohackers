package problem8

import (
	"io"
)

type decodeReader struct {
	r          io.Reader
	cypherSpec []Cypher
}

func (d *decodeReader) Read(bytes []byte) (int, error) {
	n, err := d.r.Read(bytes)
	for i := 0; i < n; i++ {
		for j := len(d.cypherSpec) - 1; j >= 0; j-- {
			cypher := d.cypherSpec[j]
			bytes[i] = cypher.operation(bytes[i], false)
		}
	}

	return n, err
}

func (d *decodeReader) resetCypherspec() {
	for _, c := range d.cypherSpec {
		c.reset()
	}
}

type encodeWriter struct {
	w          io.Writer
	cypherSpec []Cypher
}

func (e *encodeWriter) Write(bytes []byte) (int, error) {
	for i := 0; i < len(bytes); i++ {
		for _, cypher := range e.cypherSpec {
			bytes[i] = cypher.operation(bytes[i], true)
		}
	}
	n, err := e.w.Write(bytes)

	return n, err
}

func (e *encodeWriter) resetCypherspec() {
	for _, c := range e.cypherSpec {
		c.reset()
	}
}
