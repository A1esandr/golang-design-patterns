package main

import (
	"fmt"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

func sq(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			out <- n * n
		}
		close(out)
	}()
	return out
}

func merge(done <-chan struct{}, cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Запускаем output go-процедуру
	// для каждого входного канала в cs.
	// output копирует значения из c в out до тех пор,
	// пока c не закроется или не получит значение из done,
	// затем output вызывает wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			select {
			case out <- n:
			case <-done:
			}
		}
		wg.Done()
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

// Пример ранней остановки - остановка вышестоящих этапов
func main() {
	in := gen(2, 3)

	// Распределяем работу sq по двум goroutine,
	// которые обе читают из in.
	c1 := sq(in)
	c2 := sq(in)

	// Используем первое значение из вывода.
	done := make(chan struct{}, 2)
	out := merge(done, c1, c2)
	fmt.Println(<-out) // 4 или 9

	// Сообщаем оставшимся отправителям, что мы уходим.
	done <- struct{}{}
	done <- struct{}{}
}
