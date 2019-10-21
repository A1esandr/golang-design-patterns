package main

import (
	"fmt"
	"sync"
)

func gen(nums ...int) <-chan int {
	out := make(chan int)
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

func merge(cs ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)

	// Запускаем output goroutine
	// для каждого входного канала в cs.
	// output копирует значения из c в out
	// до тех пор пока c не закрыт, затем вызывает wg.Done.
	output := func(c <-chan int) {
		for n := range c {
			out <- n
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

// Пример fan in - функция merge считывает данные с нескольких входов и работает до тех пор,
// пока все они не будут закрыты, путем мультиплексирования входных каналов в один канал,
// который закрыт, когда все входы закрыты.
func main() {
	in := gen(2, 3)

	// Распределяем работу sq по двум goroutine,
	// которые обе читают из in.
	c1 := sq(in)
	c2 := sq(in)

	// Потребляем объединенный вывод из c1 и c2.
	for n := range merge(c1, c2) {
		fmt.Println(n) // 4 затем 9, или 9 затем 4
	}
}
