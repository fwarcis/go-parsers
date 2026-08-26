package csmr

import (
	"iter"
)

// Source provides lookahead and advancement over a sequence of values.
type Source[V any] interface {
	// Next advances the source by one value.
	Next()

	// Peek returns the next value without advancing the source.
	Peek() (val V, ok bool)

	// Err returns the source error, if any.
	Err() error
}

// Target accepts consumed values.
type Target[V any] interface {
	// Push appends val to the target.
	Push(val V) error
}

// Consumer matches values from a [Source] and
// optionally pushes them to a [Target].
type Consumer[V any] struct {
	src  Source[V]
	targ Target[V]

	err      error
	isFailed bool
}

// New creates a [Consumer] over src and targ.
func New[V any](
	src Source[V],
	targ Target[V],
) *Consumer[V] {
	return &Consumer[V]{
		src:  src,
		targ: targ,
	}
}

// Mode configures [Consumer] matching behavior.
type Mode uint64

// ModeDefault enables failure propagation and value pushing.
const ModeDefault Mode = 0

const (
	// ModeNotFail prevents a failed match from failing the [Consumer].
	ModeNotFail Mode = 1 << iota

	// ModeNotPush prevents matched values from being pushed to the [Target].
	ModeNotPush
)

// Exactly matches exactly n consecutive values accepted by can.
// The next value must be rejected by can or the source must be exhausted.
func (c *Consumer[V]) Exactly(
	n int,
	can func(V) bool,
	mode Mode,
) (ok bool) {
	if c.isFailed || c.err != nil {
		return false
	}

	if mode&ModeNotFail == 0 {
		defer func() {
			if !ok {
				c.isFailed = true
			}
		}()
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

// Minimum matches at least n consecutive values accepted by can.
// Matching continues until can rejects a value or the source is exhausted.
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
		defer func() {
			if !ok {
				c.isFailed = true
			}
		}()
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

// Maximum matches at most n consecutive values accepted by can.
// Fewer than n matches are valid; n+1 matching values cause failure.
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
		defer func() {
			if !ok {
				c.isFailed = true
			}
		}()
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

// ForEach matches n repetitions of sequence against consecutive source values.
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
		defer func() {
			if !ok {
				c.isFailed = true
			}
		}()
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

// Ok reports whether the [Consumer] has not failed.
func (c *Consumer[V]) Ok() bool {
	return !c.isFailed
}

// Err returns the current [Consumer] error.
func (c *Consumer[V]) Err() error {
	return c.err
}
