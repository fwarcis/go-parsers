package cns

type _runeWriter interface {
	WriteRune(rn rune) (n int, err error)
}

type RuneAdder struct {
	w _runeWriter
}

func NewRuneAdder(w _runeWriter) *RuneAdder {
	return &RuneAdder{w}
}

func (a *RuneAdder) Add(r rune) error {
	_, err := a.w.WriteRune(r)
	return err
}
