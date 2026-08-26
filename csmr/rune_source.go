package csmr

import (
	"io"
)

type runeBuffer struct {
	Rune   rune
	IsRead bool
}

type RuneSource struct {
	err error

	r   io.RuneReader
	buf runeBuffer
}

func NewRuneSource(r io.RuneReader) *RuneSource {
	p := &RuneSource{r: r}
	p.Next()

	return p
}

func (p *RuneSource) Next() {
	rn, _, err := p.r.ReadRune()
	p.err = err
	p.buf = runeBuffer{
		Rune:   rn,
		IsRead: err == nil,
	}
}

func (p *RuneSource) Peek() (rune, bool) {
	return p.buf.Rune, p.buf.IsRead
}

func (p *RuneSource) Err() error {
	return p.err
}
