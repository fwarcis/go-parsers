package csmr

type RuneWriter interface {
	WriteRune(rn rune) (n int, err error)
}

type RuneTarget struct {
	w RuneWriter
}

func NewRuneTarget(w RuneWriter) *RuneTarget {
	return &RuneTarget{w}
}

func (a *RuneTarget) Add(rn rune) error {
	_, err := a.w.WriteRune(rn)

	return err
}
