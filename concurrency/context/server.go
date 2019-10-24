// Программа server выполняет запросы в поиск Google и демонстрирует использование
// go.net Context API. Обслуживает порт 8080.
//
// /search эндпоинт принимает следующие параметры запроса:
//   q=поисковый запрос в Google
//   timeout=таймаут для запроса, в формате time.Duration
//
// Например, http://localhost:8080/search?q=golang&timeout=1s обслуживает
// несколько первых результатов поиска Google для "golang" или выдает ошибку "deadline exceeded"
// если истек таймаут.
//
// Программа представлена как пример использования Context API и не выдает настоящих результатов, поскольку
// компания Google отключила использование Google Web Search API
package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"time"

	"./google"
	"./userip"
)

func main() {
	http.HandleFunc("/search", handleSearch)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handleSearch обрабатывает запросы типа /search?q=golang&timeout=1s направляя
// запрос в google.Search. Если запрос включает timeout параметр URL-адреса,
// Context автоматически отменяется по истечении времени ожидания (timeout)
func handleSearch(w http.ResponseWriter, req *http.Request) {
	// ctx - это Context для этого обработчика. Вызов отмены закрывает
	// канал ctx.Done, который является сигналом отмены для запросов
	// запущенных этим обработчиком.
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)
	timeout, err := time.ParseDuration(req.FormValue("timeout"))
	if err == nil {
		// У запроса есть timeout, поэтому создаем контекст, который
		// автоматически отменяется по истечении времени ожидания.
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	defer cancel() // Отмена ctx, как только вернется handleSearch.

	// Проверка поискового запроса.
	query := req.FormValue("q")
	if query == "" {
		http.Error(w, "no query", http.StatusBadRequest)
		return
	}

	// Сохраняем IP-адрес пользователя в ctx
	// для использования кодом в других пакетах.
	userIP, err := userip.FromRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx = userip.NewContext(ctx, userIP)

	// Запустить поиск Google и распечатать результаты.
	start := time.Now()
	results, err := google.Search(ctx, query)
	elapsed := time.Since(start)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := resultsTemplate.Execute(w, struct {
		Results          google.Results
		Timeout, Elapsed time.Duration
	}{
		Results: results,
		Timeout: timeout,
		Elapsed: elapsed,
	}); err != nil {
		log.Print(err)
		return
	}
}

var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Title}} - <a href="{{.URL}}">{{.URL}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}; timeout {{.Timeout}}</p>
</body>
</html>
`))
