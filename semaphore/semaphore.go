package main

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNoTickets      = errors.New("не могу захватить семафор")
	ErrIllegalRelease = errors.New("не могу освободить семафор, не захватив его сначала")
)

// Semaphore содержит поведение семафора, который может быть захвачен (Acquire) и/или освобожден (Release).
type Semaphore interface {
	Acquire() error
	Release() error
}

type implementation struct {
	sem     chan struct{}
	timeout time.Duration
}

func (s *implementation) Acquire() error {
	select {
	case s.sem <- struct{}{}:
		return nil
	case <-time.After(s.timeout):
		return ErrNoTickets
	}
}

func (s *implementation) Release() error {
	select {
	case _ = <-s.sem:
		return nil
	case <-time.After(s.timeout):
		return ErrIllegalRelease
	}
}

func New(tickets int, timeout time.Duration) Semaphore {
	return &implementation{
		sem:     make(chan struct{}, tickets),
		timeout: timeout,
	}
}

func main() {
	fmt.Println("Начало")
	tickets, timeout := 1, 3*time.Second
	s := New(tickets, timeout)

	if err := s.Acquire(); err != nil {
		panic(err)
	}

	// Пробуем повторно захватить семафор
	// Захват должен потерпеть неудачу
	if err := s.Acquire(); err != nil {
		fmt.Println("Повторно захватить семафор не удалось")
		fmt.Println(err.Error())
	}

	// Выполняем важную работу

	if err := s.Release(); err != nil {
		panic(err)
	}
	fmt.Println("Завершение")
}
