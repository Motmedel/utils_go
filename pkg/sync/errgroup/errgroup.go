// Package errgroup provides synchronization, error propagation and
// context cancellation for groups of goroutines working on subtasks of a
// common task, mirroring the WithContext, Go and Wait subset of the
// golang.org/x/sync/errgroup API.
package errgroup

import (
	"context"
	"sync"
)

// Group is a collection of goroutines working on subtasks of a common
// task. The zero value is valid and does not cancel a context on error.
// A Group must not be reused for different tasks.
type Group struct {
	cancel func(error)

	waitGroup sync.WaitGroup

	errOnce sync.Once
	err     error
}

// WithContext returns a new Group and an associated context derived from
// ctx. The derived context is canceled the first time a function passed
// to Go returns a non-nil error or the first time Wait returns,
// whichever occurs first; the error is available via context.Cause.
func WithContext(ctx context.Context) (*Group, context.Context) {
	ctx, cancel := context.WithCancelCause(ctx)
	return &Group{cancel: cancel}, ctx
}

// Go calls the given function in a new goroutine. The first call to
// return a non-nil error cancels the group's context, if the group was
// created by calling WithContext; its error will be returned by Wait.
func (group *Group) Go(function func() error) {
	group.waitGroup.Go(func() {
		if err := function(); err != nil {
			group.errOnce.Do(func() {
				group.err = err
				if group.cancel != nil {
					group.cancel(group.err)
				}
			})
		}
	})
}

// Wait blocks until all function calls from the Go method have returned,
// then returns the first non-nil error (if any) from them. It cancels
// the group's context, if the group was created by calling WithContext.
func (group *Group) Wait() error {
	group.waitGroup.Wait()

	if group.cancel != nil {
		group.cancel(group.err)
	}

	return group.err
}
