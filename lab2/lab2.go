package lab2

import (
	"math"
	"sort"
	"sync"
)

type FactorizationStrategy func(base []int, from int, to int) []int

func divide(n int, m int) []int {
	result := make([]int, m)
	quotient := n / m
	remainder := n % m
	for i := 0; i < m; i++ {
		result[i] = quotient
		if remainder > 0 {
			result[i]++
			remainder--
		}
	}
	return result
}

func SieveOfEratosthenes(n int) []int {

	from := 2
	numbers := make([]int, n-from+1)
	for i := range numbers {
		numbers[i] = from + i
	}

	crossedOut := 0
	for p := from; p*p <= n; p++ {
		for i := p * p; i <= n; i += p {
			if numbers[i-from] == -1 {
				continue
			}
			numbers[i-from] = -1
			crossedOut++
		}
	}

	result := make([]int, (n-from+1)-crossedOut)
	i := 0
	for _, num := range numbers {
		if num == -1 {
			continue
		}
		result[i] = num
		i++
	}
	return result
}

func ModifiedSieveOfEratosthenes(
	n int, // * число, до которого (включительно) будут найдены все простые числа
	factorizationStrategy FactorizationStrategy,
) []int {
	nSqrt := int(math.Round(math.Sqrt(float64(n))))
	base := SieveOfEratosthenes(nSqrt)

	factorized := factorizationStrategy(base, nSqrt+1, n)

	result := make([]int, 0, len(base)+len(factorized))
	result = append(result, base...)
	result = append(result, factorized...)
	return result
}

func SequentialFactorization() FactorizationStrategy {
	return func(base []int, from int, to int) []int {
		factorized := make([]int, 0)
		for i := from; i <= to; i++ {
			isPrime := true
			for _, num := range base {
				if i%num == 0 {
					isPrime = false
					break
				}
			}
			if isPrime {
				factorized = append(factorized, i)
			}
		}
		return factorized
	}
}

type syncValue struct {
	Value int
	mtx   sync.Mutex
}

// FactorizationWithDataDecomposition — декомпозиция данных (чисел для разложения).
//
// Параметр n — количество горутин, в которых будет работать алгоритм.
func FactorizationWithDataDecomposition(threads int) FactorizationStrategy {
	return func(base []int, from int, to int) []int {

		results := make([][]int, threads)
		resultLength := syncValue{
			Value: 0,
			mtx:   sync.Mutex{},
		}

		numbersPerRoutine := divide(to-from+1, threads)

		sequentialFactorization := SequentialFactorization()

		wg := sync.WaitGroup{}
		wg.Add(threads)

		processed := from - 1
		for i := 0; i < threads; i++ {
			localFrom := processed + 1
			localTo := localFrom + numbersPerRoutine[i] - 1
			processed = localTo
			go func(
				result *[]int, from int, to int, totalLength *syncValue, wg *sync.WaitGroup,
			) {
				defer wg.Done()
				*result = sequentialFactorization(base, from, to)
				totalLength.mtx.Lock()
				totalLength.Value += len(*result)
				totalLength.mtx.Unlock()
			}(&results[i], localFrom, localTo, &resultLength, &wg)
		}

		wg.Wait()

		result := make([]int, 0, resultLength.Value)
		for _, res := range results {
			result = append(result, res...)
		}

		return result
	}
}

func FactorizationWithBaseComposition(threads int) FactorizationStrategy {
	return func(base []int, from int, to int) []int {

		factorizationResult := map[int]int{}
		resultLen := 0
		resMtx := sync.Mutex{}

		baseLen := len(base)
		baseGroupsSize := divide(baseLen, threads)

		wg := sync.WaitGroup{}
		wg.Add(threads)

		divided := 0
		for j := 0; j < threads; j++ {
			groupFrom := divided
			groupTo := groupFrom + baseGroupsSize[j]
			divided = groupTo
			group := base[groupFrom:groupTo]

			go func(base *[]int, result *map[int]int, mtx *sync.Mutex) {
				defer wg.Done()
				for i := from; i <= to; i++ {
					isPrime := true
					for _, num := range *base {
						if i%num == 0 {
							isPrime = false
							break
						}
					}
					if !isPrime {
						continue
					}

					mtx.Lock()
					factorizationResult[i]++
					if factorizationResult[i] == threads {
						resultLen++
					}
					mtx.Unlock()
				}
			}(&group, &factorizationResult, &resMtx)
		}
		wg.Wait()

		result := make([]int, resultLen)
		i := 0
		for key, value := range factorizationResult {
			if value == threads {
				result[i] = key
				i++
			}
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
		return result
	}
}

func FactorizationWithThreadPool(threads int) FactorizationStrategy {
	return func(base []int, from int, to int) []int {
		taskPool := make(chan int)

		// * Очередь задач
		go func() {
			for _, num := range base {
				taskPool <- num
			}
			close(taskPool)
		}()

		// * Функция, которая будет выполняться в потоках
		job := func(n int, resHandler func(nums []int)) {
			factorized := make([]int, 0)
			for i := from; i <= to; i++ {
				if i%n != 0 {
					factorized = append(factorized, i)
				}
			}
			resHandler(factorized)
		}

		factorizationResult := map[int]int{}
		resultLen := 0
		resMtx := sync.Mutex{}
		baseLen := len(base)

		// * Обработчик результатов из работающих потоков
		handleRes := func(nums []int) {
			resMtx.Lock()
			for _, num := range nums {
				factorizationResult[num]++
				if factorizationResult[num] == baseLen {
					resultLen++
				}
			}
			resMtx.Unlock()
		}

		wg := sync.WaitGroup{}
		wg.Add(threads)

		for t := 0; t < threads; t++ {
			// * запуск определённого количества потоков
			go func() {
				defer wg.Done()

				for {
					// * в данном потоке запускаем job каждый раз,
					// * когда канал может предоставить значение из очереди задач
					n, ok := <-taskPool
					if !ok {
						return
					}

					// * запуск job
					job(n, handleRes)
				}
			}()
		}
		wg.Wait()

		// * доп. обработка результата с последующих возвращением итогового значения
		result := make([]int, resultLen)
		i := 0
		for key, value := range factorizationResult {
			if value == baseLen {
				result[i] = key
				i++
			}
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
		return result
	}
}

type threadSafeStack struct {
	values []int
	mtx    sync.Mutex
	empty  bool
}

func initStack(nums []int) *threadSafeStack {
	stack := &threadSafeStack{}
	stack.values = nums
	stack.mtx = sync.Mutex{}
	stack.empty = len(nums) == 0
	return stack
}

func (stack *threadSafeStack) next() (int, bool) {
	stack.mtx.Lock()
	defer stack.mtx.Unlock()
	if stack.empty {
		return 0, false
	}
	length := len(stack.values)
	res := stack.values[length-1]
	stack.values = stack.values[:length-1]
	if length == 1 {
		stack.empty = true
	}
	return res, true
}

func FactorizationWithConcurrentEnumerating(threads int) FactorizationStrategy {
	return func(base []int, from int, to int) []int {

		baseStack := initStack(base)

		factorizationResult := map[int]int{}
		resultLen := 0
		resMtx := sync.Mutex{}

		baseLen := len(base)

		wg := sync.WaitGroup{}
		wg.Add(threads)

		for i := 0; i < threads; i++ {
			go func(result *map[int]int, mtx *sync.Mutex) {
				defer wg.Done()
				for {
					n, ok := baseStack.next()
					if !ok {
						return
					}
					for i := from; i <= to; i++ {
						if i%n == 0 {
							continue
						}

						mtx.Lock()
						factorizationResult[i]++
						if factorizationResult[i] == baseLen {
							resultLen++
						}
						mtx.Unlock()
					}
				}
			}(&factorizationResult, &resMtx)
		}
		wg.Wait()

		result := make([]int, resultLen)
		i := 0
		for key, value := range factorizationResult {
			if value == baseLen {
				result[i] = key
				i++
			}
		}
		sort.Slice(result, func(i, j int) bool {
			return result[i] < result[j]
		})
		return result
	}
}
