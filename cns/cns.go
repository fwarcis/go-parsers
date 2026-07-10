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
}

func New[V any](p Peeker[V], a Adder[V]) Consumer[V] {
	return Consumer[V]{p, a}
}

func (c *Consumer[V]) Exactly(n int, can func(V) bool) (bool, error) {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return false, c.p.Err()
		}
		if !can(value) {
			return false, nil
		}
		if err := c.a.Add(value); err != nil {
			return false, err
		}
		c.p.Advance()
	}
	if v, has := c.p.Peek(); has && can(v) {
		return false, nil
	}
	return true, c.p.Err()
}

func (c *Consumer[V]) Minimum(n int, can func(V) bool) (bool, error) {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return false, c.p.Err()
		}
		if !can(value) {
			return false, nil
		}
		if err := c.a.Add(value); err != nil {
			return false, err
		}
		c.p.Advance()
	}
	for {
		value, has := c.p.Peek()
		if !has {
			return true, c.p.Err()
		}
		if !can(value) {
			return true, nil
		}
		if err := c.a.Add(value); err != nil {
			return true, err
		}
		c.p.Advance()
	}
}

func (c *Consumer[V]) Maximum(n int, can func(V) bool) (bool, error) {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return true, c.p.Err()
		}
		if !can(value) {
			return true, nil
		}
		if err := c.a.Add(value); err != nil {
			return true, err
		}
		c.p.Advance()
	}
	if v, ok := c.p.Peek(); ok && can(v) {
		return false, nil
	}
	return true, c.p.Err()
}

func (c *Consumer[V]) ForEach(n int, sequence iter.Seq[V]) (bool, error) {
	for range n {
		for elem := range sequence {
			value, has := c.p.Peek()
			if !has {
				return false, c.p.Err()
			}
			if any(elem) != any(value) {
				return false, nil
			}
			if err := c.a.Add(value); err != nil {
				return false, err
			}
			c.p.Advance()
		}
	}
	return true, c.p.Err()
}
