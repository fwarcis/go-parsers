package csmr

import (
	"iter"
)

type Source[V any] interface {
	Next()
	Peek() (val V, ok bool)
	Err() error
}

type Target[V any] interface {
	Push(val V) error
}

type Consumer[V any] struct {
	src  Source[V]
	targ Target[V]

	err      error
	isFailed bool
}

func New[V any](
	src Source[V],
	targ Target[V],
) *Consumer[V] {
	return &Consumer[V]{
		src:  src,
		targ: targ,
	}
}

type Mode uint64

const ModeDefault Mode = 0

const (
	ModeNotFail Mode = 1 << iota
	ModeNotPush
)

func (c *Consumer[V]) Exactly(
	n int,
	can func(V) bool,
	mode Mode,
) (ok bool) {
	if c.isFailed || c.err != nil {
		return false
	}

	if mode&ModeNotFail == 0 {
		defer c.failIfNot(&ok)
	}

	for range n {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return false
		}

		if !can(val) {
			return false
		}

		if mode&ModeNotPush == 0 {
			c.err = c.targ.Push(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
	}

	val, has := c.src.Peek()
	if !has {
		c.err = c.src.Err()

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

	if c.err != nil {
		return n == 0
	}

	if mode&ModeNotFail == 0 {
		defer c.failIfNot(&ok)
	}

	for range n {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return false
		}

		if !can(val) {
			return false
		}

		if mode&ModeNotPush == 0 {
			c.err = c.targ.Push(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
	}

	for {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return true
		}

		if !can(val) {
			return true
		}

		if mode&ModeNotPush == 0 {
			c.err = c.targ.Push(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
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

	if c.err != nil {
		return true
	}

	if mode&ModeNotFail == 0 {
		defer c.failIfNot(&ok)
	}

	for range n {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return true
		}

		if !can(val) {
			return true
		}

		if mode&ModeNotPush == 0 {
			c.err = c.targ.Push(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
	}

	val, has := c.src.Peek()
	if !has {
		c.err = c.src.Err()

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

	if sequence == nil {
		return true
	}

	if c.err != nil {
		return false
	}

	if mode&ModeNotFail == 0 {
		defer c.failIfNot(&ok)
	}

	for range n {
		for elem := range sequence {
			val, has := c.src.Peek()
			if !has {
				c.err = c.src.Err()

				return false
			}

			if any(elem) != any(val) {
				return false
			}

			if mode&ModeNotPush == 0 {
				c.err = c.targ.Push(val)
				if c.err != nil {
					return false
				}
			}

			c.src.Next()
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
