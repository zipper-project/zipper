package db

import (
	"time"

	db "upper.io/db.v3"
)

// Struct for testing marshalling.
type timeType struct {
	// Time is handled internally as time.Time but saved as an (integer) unix
	// timestamp.
	value time.Time
}

// time.Time -> unix timestamp
func (u timeType) MarshalDB() (interface{}, error) {
	return u.value.Unix(), nil
}

// unix timestamp -> time.Time
func (u *timeType) UnmarshalDB(v interface{}) error {
	var unixTime int64

	switch t := v.(type) {
	case int64:
		unixTime = t
	case nil:
		return nil
	default:
		return db.ErrUnsupportedValue
	}

	t := time.Unix(unixTime, 0).In(time.UTC)
	*u = timeType{t}

	return nil
}

func even(i int) bool {
	if i%2 == 0 {
		return true
	}
	return false
}

func fib(i uint64) uint64 {
	if i == 0 {
		return 0
	} else if i == 1 {
		return 1
	}
	return fib(i-1) + fib(i-2)
}
