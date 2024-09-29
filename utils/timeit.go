package utils

import (
	"fmt"
	"time"
)

type measureableFunc func()

func Timeit(name string, f measureableFunc) {
	start := time.Now()

	f()

	diff := time.Since(start)

	fmt.Println(name, " took: ", diff)
}
