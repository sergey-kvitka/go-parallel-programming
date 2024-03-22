package taskengine

import (
	"sync"
	"sync/atomic"
	"time"
)

type Task func()

var tasks = []Task{nil}
var taskMtx = &sync.Mutex{}

func Consume(task Task) {
	swapped := atomic.CompareAndSwapInt32(&consumerStarted, int32(0), int32(1))
	if swapped {
		taskConsumer()
	}

	taskMtx.Lock()
	length := len(tasks)
	tasks[length-1] = task     // текущий
	tasks = append(tasks, nil) // следующий
	taskMtx.Unlock()
}

const clearCount = 100

var consumerStarted int32 = 0 // если = 1 то consumer запущен
func taskConsumer() {
	go func() { // всё происходит в одном (этом) потоке

		i := 0
		for ; ; i++ { // идём по массиву
			if i == clearCount { // обрезаем массив и уменьшаем индекс, если дошли до 100-го элемента
				taskMtx.Lock()
				tasks = tasks[clearCount:]
				i = 0
				taskMtx.Unlock()
			}

			var task Task

			for { // каждые полсекунды проверяем, не появилась ли task на проверяемом индексе
				taskMtx.Lock()
				if tasks[i] != nil {
					task = tasks[i]
					taskMtx.Unlock()
					break
				}
				taskMtx.Unlock()
				time.Sleep(time.Millisecond * 500)
			}

			task() // запуск task в том же потоке
		}

	}()
}
