package lab3

import (
	"fmt"
	"sync"
	"time"
)

func NonSync(consumerAmount int, producerMessages [][]string, logging bool) {

	var buffer *string = nil

	producerAmount := len(producerMessages)

	producers := make([]func(), producerAmount)
	consumers := make([]func(), consumerAmount)

	for p := range producers {
		i := p
		producers[i] = func() {
			messages := producerMessages[i]
			messagesAmount := len(messages)

			for m := 0; m < messagesAmount; {
				if buffer == nil {
					buffer = &messages[m]
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
				if buffer != nil {
					message := *buffer
					buffer = nil
					if logging {
						fmt.Printf("\t\t\t\t>>> Consumer %d read \"%v\"\n", i+1, message)
					}
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
