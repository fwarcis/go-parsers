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

type Result struct {
	Ok  bool
	Err error
}

func (c *Consumer[V]) Exactly(n int, can func(V) bool) Result {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return Result{false, c.p.Err()}
		}
		if !can(value) {
			return Result{false, nil}
		}
		if err := c.a.Add(value); err != nil {
			return Result{false, err}
		}
		c.p.Advance()
	}
	if v, has := c.p.Peek(); has && can(v) {
		return Result{false, nil}
	}
	return Result{true, c.p.Err()}
}

func (c *Consumer[V]) Minimum(n int, can func(V) bool) Result {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return Result{false, c.p.Err()}
		}
		if !can(value) {
			return Result{false, nil}
		}
		if err := c.a.Add(value); err != nil {
			return Result{false, err}
		}
		c.p.Advance()
	}
	for {
		value, has := c.p.Peek()
		if !has {
			return Result{true, c.p.Err()}
		}
		if !can(value) {
			return Result{true, nil}
		}
		if err := c.a.Add(value); err != nil {
			return Result{true, err}
		}
		c.p.Advance()
	}
}

func (c *Consumer[V]) Maximum(n int, can func(V) bool) Result {
	for range n {
		value, has := c.p.Peek()
		if !has {
			return Result{true, c.p.Err()}
		}
		if !can(value) {
			return Result{true, nil}
		}
		if err := c.a.Add(value); err != nil {
			return Result{true, err}
		}
		c.p.Advance()
	}
	if v, ok := c.p.Peek(); ok && can(v) {
		return Result{false, nil}
	}
	return Result{true, c.p.Err()}
}

func (c *Consumer[V]) ForEach(n int, sequence iter.Seq[V]) Result {
	for range n {
		for elem := range sequence {
			value, has := c.p.Peek()
			if !has {
				return Result{false, c.p.Err()}
			}
			if any(elem) != any(value) {
				return Result{false, nil}
			}
			if err := c.a.Add(value); err != nil {
				return Result{false, err}
			}
			c.p.Advance()
		}
	}
	return Result{true, c.p.Err()}
}
