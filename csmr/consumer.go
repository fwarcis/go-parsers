package csmr

import (
	"iter"
)

// Source provides lookahead and advancement over a sequence of values.
type Source[V any] interface {
	// Next advances the source by one value.
	Next()

	// Unscan restores the source state by undoing stepCount [Source.Next] calls.
	Unscan(stepCount int)

	// Peek returns the current value without advancing the source.
	// The value is retained by the source until [Source.Next] is called.
	//
	// - If ok is true, [Source.Err] must return nil.
	//
	// - If ok is false, [Source.Err] determines whether the source is exhausted or has failed.
	Peek() (val V, ok bool)

	// Err returns the source error.
	// If nil, the source is either healthy or exhausted.
	Err() error
}

// Target accepts consumed values.
type Target[V any] interface {
	// Add adds val to the target.
	Add(val V) error
}

// Consumer matches values from a [Source] and
// optionally adds them to a [Target].
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

const (
	// DefaultMode enables failure propagation and value addition.
	DefaultMode = Mode(0)

	// CheckMode matches values without failing the [Consumer],
	// adding them to the [Target], or advancing the [Source].
	CheckMode = ModeNotFail | ModeNotAdd | ModeUnscan

	// IgnoreMode matches and advances the [Source] without failing the [Consumer]
	// or adding values to the [Target].
	IgnoreMode = ModeNotFail | ModeNotAdd
)

const (
	// ModeNotFail prevents a failed match from failing the [Consumer].
	ModeNotFail Mode = 1 << iota

	// ModeNotAdd prevents matched values from being added to the [Target].
	ModeNotAdd

	// ModeUnscan restores the [Source] state after [Consumer] matching methods.
	ModeUnscan
)

// Exactly matches exactly n consecutive values accepted by isMatched.
// The next value must be rejected by isMatched or the source must be exhausted.
func (c *Consumer[V]) Exactly(
	n int,
	isMatched func(V) bool,
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

	stepCount := 0

	if mode&ModeUnscan != 0 {
		defer func() {
			c.src.Unscan(stepCount)
		}()
	}

	for ; stepCount < n; stepCount++ {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return false
		}

		if !isMatched(val) {
			return false
		}

		if mode&ModeNotAdd == 0 {
			c.err = c.targ.Add(val)
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

	return !isMatched(val)
}

// Minimum matches at least n consecutive values accepted by isMatched.
// Matching continues until isMatched rejects a value or the source is exhausted.
func (c *Consumer[V]) Minimum(
	n int,
	isMatched func(V) bool,
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

	stepCount := 0

	if mode&ModeUnscan != 0 {
		defer func() {
			c.src.Unscan(stepCount)
		}()
	}

	for ; stepCount < n; stepCount++ {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return false
		}

		if !isMatched(val) {
			return false
		}

		if mode&ModeNotAdd == 0 {
			c.err = c.targ.Add(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
	}

	for ; ; stepCount++ {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return true
		}

		if !isMatched(val) {
			return true
		}

		if mode&ModeNotAdd == 0 {
			c.err = c.targ.Add(val)
			if c.err != nil {
				return false
			}
		}

		c.src.Next()
	}
}

// Maximum matches at most n consecutive values accepted by isMatched.
// Fewer than n matches are valid; n+1 matching values cause failure.
func (c *Consumer[V]) Maximum(
	n int,
	isMatched func(V) bool,
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

	stepCount := 0

	if mode&ModeUnscan != 0 {
		defer func() {
			c.src.Unscan(stepCount)
		}()
	}

	for ; stepCount < n; stepCount++ {
		val, has := c.src.Peek()
		if !has {
			c.err = c.src.Err()

			return true
		}

		if !isMatched(val) {
			return true
		}

		if mode&ModeNotAdd == 0 {
			c.err = c.targ.Add(val)
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

	return !isMatched(val)
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

	stepCount := 0

	if mode&ModeUnscan != 0 {
		defer func() {
			c.src.Unscan(stepCount)
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

			if mode&ModeNotAdd == 0 {
				c.err = c.targ.Add(val)
				if c.err != nil {
					return false
				}
			}

			c.src.Next()

			stepCount++
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
