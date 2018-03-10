package design

import (
	"time"
)

type Runner interface {
	Run(distance int) (time.Duration, error)
}
