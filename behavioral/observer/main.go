// Пакет main служит примером приложения, использующего паттерн Наблюдатель (Observer).
// В песочнице play.golang.org: https://play.golang.org/p/lO8OnB73SYs
package main

import (
	"fmt"
	"time"
)

type (
	// Event определяет индикацию возникновения момента во времени.
	Event struct {
		// Data в этом случае простой int,
		// но в действительной реализации будет зависеть от приложения.
		Data int64
	}

	// Observer определяет стандартный интерфейс для экземпляров
	// быть в листе оповещения о возникновении события.
	Observer interface {
		// OnNotify позволяет событию быть "опубликованным" для реализаций интерфейса.
		// В действительных приложениях здесь также рекомендуется реализовать обработку ошибок.
		OnNotify(Event)
	}

	// Notifier это интерфейс, за экземпляром, который его реализует, будет происходить наблюдение.
	Notifier interface {
		// Register позволяет экземпляру регистрировать себя для прослушивания/наблюдения за событиями.
		Register(Observer)
		// Deregister позволяет экземпляру удалять себя из коллекции прослушиваемых/наблюдаемых.
		Deregister(Observer)
		// Notify публикует новые события для прослушивателей.
		// Метод необязателен - каждая реализация может по-своему выполнять оповещения слушателей.
		Notify(Event)
	}
)

type (
	eventObserver struct {
		id int
	}

	eventNotifier struct {
		// Использование map с пустой структурой позволяет сохранять уникальность слушателей,
		// расходуя при этом относительно мало памяти.
		observers map[Observer]struct{}
	}
)

func (o *eventObserver) OnNotify(e Event) {
	fmt.Printf("*** Наблюдатель %d получил: %d\n", o.id, e.Data)
}

func (o *eventNotifier) Register(l Observer) {
	o.observers[l] = struct{}{}
}

func (o *eventNotifier) Deregister(l Observer) {
	delete(o.observers, l)
}

func (p *eventNotifier) Notify(e Event) {
	for o := range p.observers {
		o.OnNotify(e)
	}
}

func main() {
	// Инициализируем новый Notifier.
	n := eventNotifier{
		observers: map[Observer]struct{}{},
	}

	// Регистрируем пару наблюдателей.
	n.Register(&eventObserver{id: 1})
	n.Register(&eventObserver{id: 2})

	// Простой цикл, публикующий текущий Unix timestamp наблюдателям.
	stop := time.NewTimer(10 * time.Second).C
	tick := time.NewTicker(time.Second).C
	for {
		select {
		case <-stop:
			return
		case t := <-tick:
			n.Notify(Event{Data: t.UnixNano()})
		}
	}
}
