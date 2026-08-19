package cns

import (
	"iter"
)

type Peeker[V any] interface {
	Advance()
	Peek() (val V, ok bool)
	Err() error
}

type Adder[V any] interface {
	Add(V) error
}

type Consumer[V any] struct {
	p Peeker[V]
	a Adder[V]

	err      error
	isFailed bool
}

func New[V any](p Peeker[V], a Adder[V]) Consumer[V] {
	return Consumer[V]{p, a, nil, false}
}

func (c *Consumer[V]) Exactly(
	n int,
	can func(V) bool,
) (ok bool) {
	if c.isFailed {
		return false
	}
	defer c.tryToFail(&ok)
	for range n {
		val, has := c.p.Peek()
		if !has {
			c.err = c.p.Err()
			return false
		}
		if !can(val) {
			return false
		}
		c.err = c.a.Add(val)
		if c.err != nil {
			return false
		}
		c.p.Advance()
	}
	val, has := c.p.Peek()
	if !has {
		c.err = c.p.Err()
		return true
	}
	return !can(val)
}

func (c *Consumer[V]) Minimum(n int, can func(V) bool) (ok bool) {
	if c.isFailed {
		return false
	}
	defer c.tryToFail(&ok)
	if c.err != nil {
		return false
	}
	for range n {
		val, has := c.p.Peek()
		if !has {
			c.err = c.p.Err()
			return false
		}
		if !can(val) {
			return false
		}
		c.err = c.a.Add(val)
		if c.err != nil {
			return false
		}
		c.p.Advance()
	}
	for {
		value, has := c.p.Peek()
		if !has {
			c.err = c.p.Err()
			return true
		}
		if !can(value) {
			return true
		}
		c.err = c.a.Add(value)
		if c.err != nil {
			return true
		}
		c.p.Advance()
	}
}

func (c *Consumer[V]) Maximum(n int, can func(V) bool) (ok bool) {
	if c.isFailed {
		return false
	}
	defer c.tryToFail(&ok)
	if c.err != nil {
		return true
	}
	for range n {
		val, has := c.p.Peek()
		if !has {
			c.err = c.p.Err()
			return true
		}
		if !can(val) {
			return true
		}
		c.err = c.a.Add(val)
		if c.err != nil {
			return true
		}
		c.p.Advance()
	}
	val, has := c.p.Peek()
	if !has {
		c.err = c.p.Err()
		return true
	}
	return !can(val)
}

func (c *Consumer[V]) ForEach(n int, sequence iter.Seq[V]) (ok bool) {
	if c.isFailed {
		return false
	}
	defer c.tryToFail(&ok)
	if sequence == nil {
		return true
	}
	if c.err != nil {
		return false
	}
	for range n {
		for elem := range sequence {
			val, has := c.p.Peek()
			if !has {
				c.err = c.p.Err()
				return false
			}
			if any(elem) != any(val) {
				return false
			}
			c.err = c.a.Add(val)
			if c.err != nil {
				return false
			}
			c.p.Advance()
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

func (c *Consumer[V]) tryToFail(ok *bool) {
	if !*ok {
		c.isFailed = true
	}
}
