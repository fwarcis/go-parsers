package csmr

import (
	"io"
)

type _runeBuffer struct {
	rune   rune
	isRead bool
}

type RunePeeker struct {
	err error

	r   io.RuneReader
	buf _runeBuffer
}

func NewRunePeeker(r io.RuneReader) *RunePeeker {
	p := &RunePeeker{r: r}
	p.Advance()

	return p
}

func (p *RunePeeker) Advance() {
	r, _, err := p.r.ReadRune()
	p.err = err
	p.buf = _runeBuffer{
		rune:   r,
		isRead: err == nil,
	}
}

func (p *RunePeeker) Peek() (rune, bool) {
	return p.buf.rune, p.buf.isRead
}

func (p *RunePeeker) Err() error {
	return p.err
}
