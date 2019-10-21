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
// Поскольку sq имеет одинаковый тип для входящих и исходящих каналов,
// мы можем составить его друг в друга любое количество раз.
func main() {
	// Устанавливаем пайплайн и потребляем вывод.
	for n := range sq(sq(gen(2, 3))) {
		fmt.Println(n) // 16 затем 81
	}
}
