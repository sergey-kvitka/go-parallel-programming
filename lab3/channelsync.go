package lab3

import (
	"fmt"
	"sync"
	"time"
)

func ChannelSync(consumerAmount int, producerMessages [][]string, logging bool) {

	buffer := make(chan string)

	producerAmount := len(producerMessages)

	producers := make([]func(), producerAmount)
	consumers := make([]func(), consumerAmount)

	for p := range producers {
		i := p
		producers[i] = func() {
			messages := producerMessages[i]
			messagesAmount := len(messages)

			for m := 0; m < messagesAmount; {
				buffer <- messages[m]
				if logging {
					fmt.Printf("Producer %d wrote \"%v\" >>>\n", i+1, messages[m])
				}
				m++
			}
		}
	}

	for c := range consumers {
		i := c
		consumers[i] = func() {
			for {
				message, ok := <-buffer
				if !ok {
					return
				}
				if logging {
					fmt.Printf("\t\t\t\t>>> Consumer %d read \"%v\"\n", i+1, message)
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
	close(buffer)
	consumersWg.Wait()

	elapsed := time.Since(start)
	fmt.Println("Elapsed time: " + elapsed.String())
}
