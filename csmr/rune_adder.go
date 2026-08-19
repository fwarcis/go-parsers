package csmr

type RuneWriter interface {
	WriteRune(rn rune) (n int, err error)
}

type RuneAdder struct {
	w RuneWriter
}

func NewRuneAdder(w RuneWriter) *RuneAdder {
	return &RuneAdder{w}
}

func (a *RuneAdder) Add(r rune) error {
	_, err := a.w.WriteRune(r)

	return err
}
