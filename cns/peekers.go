package cns

import (
	"io"
)

type _runeBuffer struct {
	rune   rune
	isRead bool
}

type RunePeeker struct {
	r io.RuneReader

	cur _runeBuffer
	err error
}

func NewRunePeeker(r io.RuneReader) *RunePeeker {
	p := &RunePeeker{r: r}
	p.Advance()
	return p
}

func (p *RunePeeker) Advance() {
	r, _, err := p.r.ReadRune()
	p.err = err
	p.cur = _runeBuffer{
		rune:   r,
		isRead: err == nil,
	}
}

func (p *RunePeeker) Peek() (rune, bool) {
	return p.cur.rune, p.cur.isRead
}

func (p *RunePeeker) Err() error {
	return p.err
}
