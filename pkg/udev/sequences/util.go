// Package sequences contains some utility functions releated to iter.Seq and iter.Seq2
package sequences

import "iter"

// Unfold creates a iterator by unfolding a object T. When calling the iterator,
// init() is called to fetch the initial object which is yielded. After yielding next() is
// called with the current object and returns the next object and whether this object is valid.
// If the object is invalid (ok == false), the iterator stops.
func Unfold[T any](init func() T, next func(T) (T, bool)) iter.Seq[T] {
	return func(yield func(T) bool) {
		obj := init()
		for {
			if !yield(obj) {
				return
			}
			nextobj, ok := next(obj)
			if !ok {
				return
			}
			obj = nextobj
		}
	}
}

// Map takes an input single-value iterator and calls cb() for every item. The result of cb() is yielded to the output single-value iterator.
func Map[F, T any](input iter.Seq[F], cb func(F) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		input(func(item F) bool {
			return yield(cb(item))
		})
	}
}

// Map2 takes an input 2-value iterator and calls cb() for every item. The result of cb() is yielded to the output 2-value iterator.
func Map2[F1, F2, T1, T2 any](input iter.Seq2[F1, F2], cb func(F1, F2) (T1, T2)) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		input(func(first F1, second F2) bool {
			return yield(cb(first, second))
		})
	}
}

// Map12 takes an input single-value iterator and calls cb() for every item. The result of cb() is yielded to the output 2-value iterator.
func Map12[F, T1, T2 any](input iter.Seq[F], cb func(F) (T1, T2)) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		input(func(item F) bool {
			return yield(cb(item))
		})
	}
}

// Map12 takes an input 2-value iterator and calls cb() for every item. The result of cb() is yielded to the output single-value iterator.
func Map21[F1, F2, T any](input iter.Seq2[F1, F2], cb func(F1, F2) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		input(func(first F1, second F2) bool {
			return yield(cb(first, second))
		})
	}
}

// Filter takes an input single-value iterator and calls test() for every item.
// If the result is positive, this item is yielded to the output single-value iterator.
func Filter[T any](input iter.Seq[T], test func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		input(func(item T) bool {
			if !test(item) {
				// don't yield but do continue
				return true
			}
			return yield(item)
		})
	}
}

// Filter2 takes an input 2-value iterator and calls test() for every item.
// If the result is positive, this item is yielded to the output 2-value iterator.
func Filter2[T1, T2 any](input iter.Seq2[T1, T2], test func(T1, T2) bool) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		input(func(first T1, second T2) bool {
			if !test(first, second) {
				// don't yield but do continue
				return true
			}
			return yield(first, second)
		})
	}
}

// Enumerate takes an input single-value iterator and returns an 2-value iterator combining
// the index of this iterator (zero-based) and the item.
func Enumerate[T any](input iter.Seq[T]) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := 0
		input(func(item T) bool {
			cont := yield(i, item)
			i++
			return cont
		})
	}
}
