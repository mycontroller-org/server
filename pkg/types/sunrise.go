package types

import "time"

type Sunrise interface {
	SunriseTime() (*time.Time, error)
	SunsetTime() (*time.Time, error)
}
