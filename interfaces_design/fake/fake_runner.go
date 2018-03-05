package fake

import (
	"sync"
	"time"
)

type Runner struct {
	runMethod map[int]RunnerRunMethod
	runMutex  sync.RWMutex
}

type RunnerRunMethod struct {
	DistanceArg int

	Called     bool
	TimeResult time.Duration
	ErrResult  error
}

func (method RunnerRunMethod) WasCalled() bool { return method.Called }

type RunnerRunFunc func(RunnerRunMethod) RunnerRunMethod

func NewRunner() *Runner {
	fake := &Runner{}
	fake.runMethod = make(map[int]RunnerRunMethod)

	return fake
}

func (fake *Runner) RunReturns(timeResult time.Duration, errResult error) *Runner {
	fake.runMutex.Lock()
	fakeMethod := fake.runMethod[0]
	fakeMethod.TimeResult = timeResult
	fakeMethod.ErrResult = errResult
	fake.runMethod[0] = fakeMethod
	fake.runMutex.Unlock()

	return fake
}

func (fake *Runner) RunGetArgs() (distanceArg int) {
	fake.runMutex.RLock()
	distanceArg = fake.runMethod[0].DistanceArg
	fake.runMutex.RUnlock()

	return distanceArg
}

func (fake *Runner) RunForCall(call int, fns ...RunnerRunFunc) *Runner {
	fake.runMutex.Lock()
	for _, fn := range fns {
		fakeMethod := fake.runMethod[call]
		fake.runMethod[call] = fn(fakeMethod)
	}
	fake.runMutex.Unlock()

	return fake
}
