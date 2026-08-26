package csmr

import (
	"errors"
	"io"
)

type runeBuffer struct {
	Rune   rune
	IsRead bool
}

type RuneSource struct {
	err error

	scanr io.RuneScanner
	buf   runeBuffer
}

func NewRuneSource(scanr io.RuneScanner) *RuneSource {
	p := &RuneSource{scanr: scanr}
	p.Next()

	return p
}

func (p *RuneSource) Next() {
	rn, _, err := p.scanr.ReadRune()
	p.err = err
	p.buf = runeBuffer{
		Rune:   rn,
		IsRead: err == nil,
	}
}

func (p *RuneSource) Unscan(stepCount int) {
	if stepCount <= 0 {
		return
	}

	p.err = p.scanr.UnreadRune()

	for range stepCount - 1 {
		p.err = errors.Join(
			p.err, p.scanr.UnreadRune(),
		)
	}
}

func (p *RuneSource) Peek() (rune, bool) {
	return p.buf.Rune, p.buf.IsRead
}

func (p *RuneSource) Err() error {
	return p.err
}
