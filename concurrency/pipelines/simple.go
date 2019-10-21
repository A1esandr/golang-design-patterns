package main

import (
	"fmt"
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

// Пример простого пайплайна из 3 этапов
// 1 этап - gen - принимает значения и передает
// 2 этап - sq - принимает значения, обрабатывает и передает результат
// 3 этап - main - принимает результаты и использует
func main() {
	// Устанавливаем пайплайн.
	c := gen(2, 3)
	out := sq(c)

	// Потребляем вывод.
	fmt.Println(<-out) // 4
	fmt.Println(<-out) // 9
}
