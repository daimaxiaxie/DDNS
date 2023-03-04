package dnsutil

import "fmt"

type Constraint interface {
	comparable
	fmt.Stringer
}

func Equal[T comparable](value, expect T) bool {
	if value == expect {
		return true
	}
	return false
}

func Unequal[T comparable](value, expect T) bool {
	if value == expect {
		return false
	}
	return true
}

func MustEqual[T comparable](value, expect T) {
	if value != expect {
		panic("value not allowed")
	}
}

func MustUnequal[T comparable](value, expect T) {
	if value == expect {
		panic("value not allowed")
	}
}
