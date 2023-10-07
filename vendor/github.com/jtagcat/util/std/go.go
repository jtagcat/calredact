package std

import "sync"

// go <func> with single waitgroup
//
//	waitGo := util.GoWg(func() {
//		exampleFunc(foo, bar)
//	})
//	defer waitGo()
func GoWg(fn func()) (wait func()) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()

	return wg.Wait
}
