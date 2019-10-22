package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// walkFiles запускает goroutine для обхода дерева каталогов в root каталоге
// и отправки пути каждого обычного файла по string каналу.
// Отправляет результат walk по каналу ошибок.
// Если done закрыт, walkFiles прекращает свою работу.
func walkFiles(done <-chan struct{}, root string) (<-chan string, <-chan error) {
	paths := make(chan string)
	errc := make(chan error, 1)
	go func() {
		// Закрываем paths канал после возврата Walk.
		defer close(paths)
		// select не нужен здесь, поскольку errc буферизован.
		errc <- filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			select {
			case paths <- path:
			case <-done:
				return errors.New("walk canceled")
			}
			return nil
		})
	}()
	return paths, errc
}

// result - продукт чтения и суммирования файла с использованием MD5.
type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

// digester считывает имена путей из paths и отправляет дайджесты соответствующих файлов по c,
// пока либо paths, либо done не будут закрыты.
func digester(done <-chan struct{}, paths <-chan string, c chan<- result) {
	for path := range paths {
		data, err := ioutil.ReadFile(path)
		select {
		case c <- result{path, md5.Sum(data), err}:
		case <-done:
			return
		}
	}
}

// MD5All читает все файлы в дереве файлов с корнем в root и возвращает карту
// пути к файлу к MD5 сумме содержимого файла.
// Если происходит сбой прохода по каталогу
// или сбой любой операции чтения, MD5All возвращает ошибку.
// В этом случае MD5All не ожидает завершения выполняющихся операций чтения.
func MD5All(root string) (map[string][md5.Size]byte, error) {
	// MD5All закрывает done канал при возврате;
	// это может быть сделано
	// до получения всех значений от c и errc.
	done := make(chan struct{})
	defer close(done)

	paths, errc := walkFiles(done, root)

	// Запускаем фиксированное количество go-процедур для чтения и получения дайджеста файлов.
	c := make(chan result)
	var wg sync.WaitGroup
	const numDigesters = 20
	wg.Add(numDigesters)
	for i := 0; i < numDigesters; i++ {
		go func() {
			digester(done, paths, c)
			wg.Done()
		}()
	}
	go func() {
		wg.Wait()
		close(c)
	}()
	// Конец пайплайна

	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}
	// Проверяем не произошел ли сбой в Walk.
	if err := <-errc; err != nil {
		return nil, err
	}
	return m, nil
}

func main() {
	// Рассчитать MD5 сумму всех файлов
	// в указанном каталоге,
	// затем печатаем результаты,
	// отсортированные по имени пути.
	m, err := MD5All(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	var paths []string
	for path := range m {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		fmt.Printf("%x  %s\n", m[path], path)
	}
}
