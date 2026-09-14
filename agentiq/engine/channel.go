package engine

import "fmt"

var (
	ErrMultipleWrites = fmt.Errorf("multiple writes to channel")
)

type Channel interface {
	Updates(writes []any) (changed bool, err error)
	Get() any
	Peek(writes []any) any
}

type LastValue struct {
	value any
}

func NewLastValue(initial any) *LastValue {
	return &LastValue{value: initial}
}

// Get implements [Channel].
func (c *LastValue) Get() any {
	return c.value
}

// Peek implements [Channel].
func (c *LastValue) Peek(writes []any) any {
	if len(writes) == 0 {
		return c.value
	}
	return writes[len(writes)-1]
}

// Updates implements [Channel].
func (c *LastValue) Updates(writes []any) (changed bool, err error) {
	switch len(writes) {
	case 0:
		return false, nil
	case 1:
		c.value = writes[0]
		return true, nil
	default:
		return false, fmt.Errorf("%w, %d writes in one tick", ErrMultipleWrites, len(writes))
	}
}

type BinaryOperatorAggregate struct {
	op  func(acc, next any) any
	val any
}


// Get implements [Channel].
func (c *BinaryOperatorAggregate) Get() any {
	return c.val
}

// Peek implements [Channel].
func (c *BinaryOperatorAggregate) Peek(writes []any) any {
	v := c.val
	for _, w := range writes {
		v = c.op(v, w)
	}
	return v
}

// Updates implements [Channel].
func (c *BinaryOperatorAggregate) Updates(writes []any) (changed bool, err error) {
	if len(writes) == 0 {
		return false, nil
	}
	for _, w := range writes {
		c.val = c.op(c.val, w)
	}
	return true, nil
}

func NewAggregate(initial any, op func(acc, next any) any) *BinaryOperatorAggregate {
	return &BinaryOperatorAggregate{
		op:  op,
		val: initial,
	}
}

func NewAppend() *BinaryOperatorAggregate {
	return NewAggregate([]any{}, func(acc, next any) any {
		old := acc.([]any)
		out := make([]any, len(old), len(old)+1)
		copy(out, old)
		return append(out, next)
	})
}

func NewSum() *BinaryOperatorAggregate {
	return NewAggregate(0, func(acc, next any) any {
		return acc.(int) + next.(int)
	})
}

var _ Channel = (*LastValue)(nil)
var _ Channel = (*BinaryOperatorAggregate)(nil)
