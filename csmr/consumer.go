package csmr

import (
	"iter"
)

type Peeker[V any] interface {
	Advance()
	Peek() (val V, ok bool)
	Err() error
}

type Adder[V any] interface {
	Add(val V) error
}

type Consumer[V any] struct {
	pkr Peeker[V]
	adr Adder[V]

	err      error
	isFailed bool
}

func New[V any](pkr Peeker[V], adr Adder[V]) *Consumer[V] {
	return &Consumer[V]{
		pkr: pkr,
		adr: adr,
	}
}

type Mode uint64

const (
	ModeCheck Mode = 1 << iota

	ModeDefault Mode = 0
)

func (c *Consumer[V]) Exactly(
	n int,
	can func(V) bool,
	mode Mode,
) (ok bool) {
	if c.isFailed {
		return false
	}

	if mode&ModeCheck == 0 {
		defer c.failIfNot(&ok)
	}

	for range n {
		val, has := c.pkr.Peek()
		if !has {
			c.err = c.pkr.Err()

			return false
		}

		if !can(val) {
			return false
		}

		c.err = c.adr.Add(val)
		if c.err != nil {
			return false
		}

		c.pkr.Advance()
	}

	val, has := c.pkr.Peek()
	if !has {
		c.err = c.pkr.Err()

		return true
	}

	return !can(val)
}

func (c *Consumer[V]) Minimum(
	n int,
	can func(V) bool,
	mode Mode,
) (ok bool) {
	if c.isFailed {
		return false
	}

	if mode&ModeCheck == 0 {
		defer c.failIfNot(&ok)
	}

	if c.err != nil {
		return false
	}

	for range n {
		val, has := c.pkr.Peek()
		if !has {
			c.err = c.pkr.Err()

			return false
		}

		if !can(val) {
			return false
		}

		c.err = c.adr.Add(val)
		if c.err != nil {
			return false
		}

		c.pkr.Advance()
	}

	for {
		value, has := c.pkr.Peek()
		if !has {
			c.err = c.pkr.Err()

			return true
		}

		if !can(value) {
			return true
		}

		c.err = c.adr.Add(value)
		if c.err != nil {
			return true
		}

		c.pkr.Advance()
	}
}

func (c *Consumer[V]) Maximum(
	n int,
	can func(V) bool,
	mode Mode,
) (ok bool) {
	if c.isFailed {
		return false
	}

	if mode&ModeCheck == 0 {
		defer c.failIfNot(&ok)
	}

	if c.err != nil {
		return true
	}

	for range n {
		val, has := c.pkr.Peek()
		if !has {
			c.err = c.pkr.Err()

			return true
		}

		if !can(val) {
			return true
		}

		c.err = c.adr.Add(val)
		if c.err != nil {
			return true
		}

		c.pkr.Advance()
	}

	val, has := c.pkr.Peek()
	if !has {
		c.err = c.pkr.Err()

		return true
	}

	return !can(val)
}

func (c *Consumer[V]) ForEach(
	n int,
	sequence iter.Seq[V],
	mode Mode,
) (ok bool) {
	if c.isFailed {
		return false
	}

	if mode&ModeCheck == 0 {
		defer c.failIfNot(&ok)
	}

	if sequence == nil {
		return true
	}

	if c.err != nil {
		return false
	}

	for range n {
		for elem := range sequence {
			val, has := c.pkr.Peek()
			if !has {
				c.err = c.pkr.Err()

				return false
			}

			if any(elem) != any(val) {
				return false
			}

			c.err = c.adr.Add(val)
			if c.err != nil {
				return false
			}

			c.pkr.Advance()
		}
	}

	return true
}

func (c *Consumer[V]) Ok() bool {
	return !c.isFailed
}

func (c *Consumer[V]) Err() error {
	return c.err
}

func (c *Consumer[V]) failIfNot(ok *bool) {
	if !*ok {
		c.isFailed = true
	}
}
