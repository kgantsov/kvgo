package kv

import (
	"sync/atomic"
)

type Bool struct {
	value int32
}

func NewBool() Bool {
	return Bool{value: 0}
}

func (b *Bool) Value() bool {
	if atomic.LoadInt32(&b.value) == 1 {
		return true
	} else {
		return false
	}
}

func (b *Bool) Set(value bool) {
	if value {
		atomic.StoreInt32(&b.value, 1)
	} else {
		atomic.StoreInt32(&b.value, 0)
	}
}
