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

// Результатом является продукт чтения и суммирования файла с использованием MD5.
type result struct {
	path string
	sum  [md5.Size]byte
	err  error
}

// sumFiles запускает go-процедуры для обхода дерева каталогов в root
// и получения дайджеста каждого обычного файла. Эти go-процедуры отправляют
// результаты дайджестов по каналу result и отправляют результат обхода по каналу ошибок.
// Если done закрыт, sumFiles прекращает свою работу.
func sumFiles(done <-chan struct{}, root string) (<-chan result, <-chan error) {
	// Для каждого обычного файла запускаем goroutine, которая суммирует файл и отправляет
	// результат в c. Ошибки walk отправляются в errc.
	c := make(chan result)
	errc := make(chan error, 1)
	go func() {
		var wg sync.WaitGroup
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.Mode().IsRegular() {
				return nil
			}
			wg.Add(1)
			go func() {
				data, err := ioutil.ReadFile(path)
				select {
				case c <- result{path, md5.Sum(data), err}:
				case <-done:
				}
				wg.Done()
			}()
			// Завершаем walk если done закрыт.
			select {
			case <-done:
				return errors.New("walk canceled")
			default:
				return nil
			}
		})
		// Walk вернулся, поэтому все вызовы wg.Add завершены.
		// Начинаем goroutine для закрытия c,
		// как только все посылки сделаны.
		go func() {
			wg.Wait()
			close(c)
		}()
		// select не нужен здесь, поскольку errc буферизован.
		errc <- err
	}()
	return c, errc
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

	c, errc := sumFiles(done, root)

	m := make(map[string][md5.Size]byte)
	for r := range c {
		if r.err != nil {
			return nil, r.err
		}
		m[r.path] = r.sum
	}
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
