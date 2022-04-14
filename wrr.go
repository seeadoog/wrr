package wrr

import (
	"fmt"
	"math"
	"sync"
)

type Target[T any] struct {
	Weight int
	Target T
}

type target[T any] struct {
	weight          int
	currentWeight   int
	effectiveWeight int
	t               T
}

type WrrLooper[T any] struct {
	targets       []*target[T]
	lock          sync.RWMutex
	defaultWeight int
	deltaWeight   int
}

func NewWrrLooper[T any](deltaWeight int, defaultWeight int) *WrrLooper[T] {
	return &WrrLooper[T]{
		defaultWeight: defaultWeight,
		deltaWeight:   defaultWeight,
	}
}

func (w *WrrLooper[T]) SetTargets(tgs ...*Target[T]) {

	targets := ConvertSlice(tgs, func(i *Target[T]) *target[T] {
		return &target[T]{
			weight:          i.Weight,
			currentWeight:   i.Weight,
			effectiveWeight: i.Weight,
			t:               i.Target,
		}
	})
	w.lock.Lock()
	w.targets = targets
	w.lock.Unlock()
}

func (w *WrrLooper[T]) GetTargets() (res []T) {
	w.lock.RLock()
	res = ConvertSlice(w.targets, func(i *target[T]) T {
		return i.t
	})
	w.lock.RUnlock()
	return
}

func (w *WrrLooper[T]) selectTarget() *target[T] {
	if len(w.targets) == 0 {
		return nil
	}
	total := 0
	var target *target[T]
	max := math.MinInt64
	w.lock.Lock()
	for _, t := range w.targets {
		t.currentWeight += t.effectiveWeight
		total += t.effectiveWeight
		if t.currentWeight > max {
			max = t.currentWeight
			target = t
		}
	}
	target.currentWeight -= total
	w.lock.Unlock()
	return target
}

func (w *WrrLooper[T]) Call(f func(T) error) (err error) {
	t := w.selectTarget()
	if t == nil {
		return fmt.Errorf("No peer Addr found")
	}
	err = f(t.t)
	if err != nil {
		if w.deltaWeight > 0 {
			w.lock.Lock()
			if t.effectiveWeight-w.deltaWeight > 0 {
				t.effectiveWeight = t.effectiveWeight - w.deltaWeight
			} else {
				t.effectiveWeight = 1
			}
			w.lock.Unlock()
		}

	} else {
		if t.effectiveWeight < t.weight && w.deltaWeight > 0 {
			w.lock.Lock()
			if t.effectiveWeight < t.weight {
				if t.effectiveWeight+w.deltaWeight <= t.weight {
					t.effectiveWeight = t.effectiveWeight + w.deltaWeight
				} else {
					t.effectiveWeight = t.weight
				}
			}
			w.lock.Unlock()

		}
	}
	return err
}

func ConvertSlice[I any, O any](in []I, f func(i I) O) (out []O) {
	out = make([]O, len(in))
	for i, e := range in {
		o := f(e)
		out[i] = o
	}
	return out
}
