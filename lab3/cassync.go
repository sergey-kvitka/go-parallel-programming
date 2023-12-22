package lab3

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func CasSync(consumerAmount int, producerMessages [][]int, logging bool) {

	var buffer int32 = 0 // 0 будет означать, что буфер пустой

	producerAmount := len(producerMessages)

	producers := make([]func(), producerAmount)
	consumers := make([]func(), consumerAmount)

	for p := range producers {
		i := p
		producers[i] = func() {
			messages := producerMessages[i]
			messagesAmount := len(messages)

			for m := 0; m < messagesAmount; {

				value := buffer
				if value == 0 && atomic.CompareAndSwapInt32(&buffer, value, int32(messages[m])) {
					if logging {
						fmt.Printf("Producer %d wrote \"%v\" >>>\n", i+1, messages[m])
					}
					m++
				}
			}
		}
	}

	end := false

	for c := range consumers {
		i := c
		consumers[i] = func() {
			for {
				value := atomic.SwapInt32(&buffer, 0)
				if value != 0 && logging {
					fmt.Printf("\t\t\t\t>>> Consumer %d read \"%v\"\n", i+1, value)
				}

				if end {
					return
				}
			}
		}
	}

	consumersWg := sync.WaitGroup{}
	consumersWg.Add(consumerAmount)

	producersWg := sync.WaitGroup{}
	producersWg.Add(producerAmount)

	start := time.Now()

	for _, c := range consumers {
		consumer := c
		go func() {
			defer consumersWg.Done()
			consumer()
		}()
	}

	for _, p := range producers {
		producer := p
		go func() {
			defer producersWg.Done()
			producer()
		}()
	}

	producersWg.Wait()
	end = true
	consumersWg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Elapsed time: " + elapsed.String())
}
