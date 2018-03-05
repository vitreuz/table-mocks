package design

import (
	"testing"
	"time"

	"github.com/vitreuz/table-mocks/interfaces_design/fake"
)

type Runner interface {
	Run(distance int) (time.Duration, error)
}

type Cat struct {
	speed int
}

// We'll make dummy function that tests which cat wins a race. This is
// determined by the cat which completes the distance in the fastest time. Each
// cat will be a struct with a speed element to it
func TestCat(t *testing.T) {
	type checkOut func(Cat) []error

	runReturns := func(time time.Duration, err error) fake.RunnerRunFunc {
		return func(method fake.RunnerRunMethod) fake.RunnerRunMethod {
			method.TimeResult = time
			method.ErrResult = err
			return method
		}
	}

	tests := [...]struct {
		name       string
		fakeRunner *fake.Runner
		checks     []checkOut
	}{
		{
			"some foo test",
			fake.NewRunner().
				RunForCall(1, runReturns(1*time.Second, nil)).
				RunForCall(2, runReturns(2*time.Second, nil)),
			nil,
		},
	}

	_ = tests
}
