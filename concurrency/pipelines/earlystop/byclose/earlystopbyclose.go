package byclose

import (
	"fmt"
	"sync"
)

func gen(done <-chan struct{}, nums ...int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		for _, n := range nums {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
		close(out)
	}()
	return out
}

func sq(done <-chan struct{}, in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			select {
			case out <- n * n:
			case <-done:
				return
			}
		}
	}()
	return out
}

func merge(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Запускаем output goroutine
	// для каждого входного канала в cs.
	// output копирует значения из c в out
	// до закрытия c или done,
	// затем вызывает wg.Done.
	output := func(c <-chan int) {
		defer wg.Done()
		for n := range c {
			select {
			case out <- n:
			case <-done:
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Запускаем goroutine чтобы закрыть out
	// когда все output goroutine заверешены.
	// Это должно начнаться после вызова wg.Add.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// Пример ранней остановки вышестоящих этапов, посредством закрытия канала
func main() {
	// Установливаем done канал, общий для всего пайплайна,
	// и закрываем этот канал при выходе из этого пайплайна
	// в качестве сигнала для всех go-процедур,
	// что мы начали выходить.
	done := make(chan struct{})
	defer close(done)

	in := gen(done, 2, 3)

	// Распределяем работу sq по двум goroutine,
	// которые обе читают из in.
	c1 := sq(done, in)
	c2 := sq(done, in)

	// Используем первое значение из output.
	out := merge(done, c1, c2)
	fmt.Println(<-out) // 4 or 9

	// done будет закрыт отложенным вызовом.
}
