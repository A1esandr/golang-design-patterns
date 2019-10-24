// Пакет google предоставляет функцию для отправки запросов в Google,
// используя Google Web Search API.
//
// Этот пакет представлен для примера работы context
package google

import (
	"context"
	"encoding/json"
	"net/http"

	"../userip"
)

// Results это упорядоченный список результатов поиска.
type Results []Result

// Result содержит заголовок и URL результата поиска.
type Result struct {
	Title, URL string
}

// Search отправляет запрос в поиск Google и возвращает результаты.
func Search(ctx context.Context, query string) (Results, error) {
	// Подготовливаем запрос API поиска Google.
	req, err := http.NewRequest("GET", "https://ajax.googleapis.com/ajax/services/search/web?v=1.0", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("q", query)

	// Если ctx передает IP-адрес пользователя, перенаправляем его на сервер.
	// API Google используют IP-адрес пользователя для различения запросов,
	// инициированных сервером от запросов конечного пользователя.
	if userIP, ok := userip.FromContext(ctx); ok {
		q.Set("userip", userIP.String())
	}
	req.URL.RawQuery = q.Encode()

	// Выполняем HTTP запрос и обрабатываем ответ.
	// Функция httpDo отменяет запрос если ctx.Done закрыт.
	var results Results
	err = httpDo(ctx, req, func(resp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Обрабатываем JSON результат поиска.
		var data struct {
			ResponseData struct {
				Results []struct {
					TitleNoFormatting string
					URL               string
				}
			}
		}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}
		for _, res := range data.ResponseData.Results {
			results = append(results, Result{Title: res.TitleNoFormatting, URL: res.URL})
		}
		return nil
	})
	// httpDo ожидает возврата из предоставленного нами замыкания,
	// поэтому безопасно читать результаты здесь.
	return results, err
}

// httpDo выполняет HTTP запрос и вызывает f с ответом.
// Если ctx.Done закрыт, пока запрос или f выполняется, httpDo отменяет запрос,
// ожидает выхода f, и возвращает ctx.Err. Иначе, httpDo возвращает ошибку от f.
func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	// Запускаем HTTP-запрос в goroutine и передаем ответ в f.
	c := make(chan error, 1)
	req = req.WithContext(ctx)
	go func() { c <- f(http.DefaultClient.Do(req)) }()
	select {
	case <-ctx.Done():
		<-c // Ожидаем пока f вернется.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
