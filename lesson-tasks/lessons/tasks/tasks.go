package tasks

import (
	"errors"
	"sync"
)

func ParallelSliceSumChan(nums []int) int {
	var (
		length  = len(nums)
		sumChan = make(chan int)
	)
	if length == 0 {
		return 0
	}
	go countSumChan(nums, 0, length/2, sumChan)
	go countSumChan(nums, length/2, length, sumChan)
	return <-sumChan + <-sumChan
}

func countSumChan(nums []int, from int, to int, sumChan chan int) {
	var sum int
	for i := from; i < to; i++ {
		sum += nums[i]
	}
	if sum == 0 {
		return
	}
	sumChan <- sum
}

type sumMtx struct {
	mtx *sync.Mutex
	sum int
}

func ParallelSliceSumWg(nums []int) int {
	var (
		wg        = &sync.WaitGroup{}
		sumWriter = &sumMtx{mtx: &sync.Mutex{}}
		length    = len(nums)
	)

	if length == 0 {
		return 0
	}

	wg.Add(2)

	go countSumWg(nums, 0, length/2, sumWriter, wg)
	go countSumWg(nums, length/2, length, sumWriter, wg)

	wg.Wait()

	return sumWriter.sum
}

func countSumWg(nums []int, from int, to int, sumWriter *sumMtx, wg *sync.WaitGroup) {
	defer wg.Done()
	var sum int
	for i := from; i < to; i++ {
		sum += nums[i]
	}
	if sum == 0 {
		return
	}
	sumWriter.mtx.Lock()
	sumWriter.sum += sum
	sumWriter.mtx.Unlock()
}

func FindUnique(numbers []int) (int, error) {

	repeats := make(map[int]int)

	for _, value := range numbers {
		repeats[value]++
	}

	for key, value := range repeats {
		if value == 1 {
			return key, nil
		}
	}
	return 0, errors.New("в срезе нет уникального числа")
}

func Fibonacci(number int) int {
	prev := 0
	if number == 1 {
		return prev
	}
	cur := 1

	for i := 3; i <= number; i++ {
		prev, cur = cur, cur+prev
	}
	return cur
}

func GetMax(slice []int) (int, int) {
	max := slice[0]
	var maxIndex int

	for index, value := range slice {
		if value > max {
			maxIndex = index
			max = value
		}
	}
	return max, maxIndex
}
