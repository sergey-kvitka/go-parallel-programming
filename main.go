package main

import (
	"flag"
	"fmt"
	"go-parallel-programming/lab1"
	"go-parallel-programming/lab2"
	"go-parallel-programming/lab3"
	"math"
	"runtime"
	"time"
)

func main() {

}

func lab3Test() {
	consumers := 3
	messages := lab3.GenerateMessages(4, 6)

	lab3.NonSync(consumers, messages, true)
	fmt.Print("\n\n")
	lab3.MutexSync(consumers, messages, true)
	fmt.Print("\n\n")
	lab3.ChannelSync(consumers, messages, true)
	fmt.Print("\n\n")
	lab3.CasSync(consumers, lab3.GenerateIntMessages(4, 6), true)
}

func lab2Test() {

	printSliceEdges := func(nums []int) {
		length := len(nums)
		if length <= 8 {
			fmt.Println(nums)
			return
		}
		min := int(math.Min(math.Floor(float64(length)/2), 6))
		start := fmt.Sprint(nums[0 : min-1])
		end := fmt.Sprint(nums[length-min+1 : length])
		fmt.Println(start[:len(start)-1], "...", end[1:])
	}

	threads := runtime.NumCPU()
	fmt.Println("Количество потоков:", threads)

	_n := 1e6
	n := int(_n)
	fmt.Println("Поиск простых чисел в диапазоне от 2 до", n)

	_time := time.Now()

	fmt.Println("\nАлгоритм нахождения базовых простых чисел: решето Эратосфена")
	printSliceEdges(lab2.SieveOfEratosthenes(n))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	_time = time.Now()

	fmt.Println("\nМодифицированный последовательный алгоритм")
	printSliceEdges(lab2.ModifiedSieveOfEratosthenes(n, lab2.SequentialFactorization()))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	fmt.Println(time.Since(_time))
	_time = time.Now()

	fmt.Println("\nПараллельный алгоритм №1: декомпозиция по данным")
	printSliceEdges(lab2.ModifiedSieveOfEratosthenes(n, lab2.FactorizationWithDataDecomposition(threads)))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	fmt.Println(time.Since(_time))
	_time = time.Now()

	fmt.Println("\nПараллельный алгоритм №2: декомпозиция по базовым простым числам")
	printSliceEdges(lab2.ModifiedSieveOfEratosthenes(n, lab2.FactorizationWithBaseComposition(threads)))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	fmt.Println(time.Since(_time))
	_time = time.Now()

	fmt.Println("\nПараллельный алгоритм №3: применение пула потоков")
	printSliceEdges(lab2.ModifiedSieveOfEratosthenes(n, lab2.FactorizationWithThreadPool(threads)))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	fmt.Println(time.Since(_time))
	_time = time.Now()

	fmt.Println("\nПараллельный алгоритм №4: последовательный перебор простых чисел")
	printSliceEdges(lab2.ModifiedSieveOfEratosthenes(n, lab2.FactorizationWithConcurrentEnumerating(threads)))
	fmt.Println("Время работы данного алгоритма:", time.Since(_time))
	fmt.Println(time.Since(_time))
	_time = time.Now()
}

// lab1TestFlags запускает обработку N чисел в M горутинах.
//
// Аргументы для работы функции берутся из флагов из командной строки, которые передаются при запуске программы.
//
// • Параметр N (флаг: --N, 10 ≤ N ≤ 1,000,000,000, пример: --N=10000) задаёт количество чисел, которые будут записаны
// в файл, и далее считаны для обработки;
//
// • Параметр M (флаг: --M, 1 ≤ M ≤ 1000, M ≤ N, пример: --M=30) задаёт количество горутин, в которых будут
// обрабатываться числа из файла;
//
// • Параметр fdistr (флаг --fdistr, fdistr ∈ ["const", "hyp"], пример: --fdistr=const) обозначает один из двух
// видов распределения чисел по потокам — равномерное распределение ("const") и распределение по гиперболе ("hyp");
//
// • Параметр fproc (флаг --fproc, fproc ∈ ["mult2", "pow2"], пример: --fproc=pow2) обозначает один из двух
// видов операции над отдельными числами — умножение на 2 ("mult2") и возведение в квадрат ("pow2")
//
// Пример запуска программы, функция main которой вызывает данный метод:
//  go run main.go --N=1000000 --M=100 --fdistr=const --fproc=pow2
// (обработка 1 000 000 чисел, равномерно распределённых по 100 горутинам, путём возведения чисел в квадрат)
func lab1TestFlags() {
	// определение обрабатываемых флагов (названия, значения по умолчанию и описания)
	N := flag.Int("N", 10, "количество чисел для обработки")
	M := flag.Int("M", 1, "количество горутин для обработки чисел")
	distrFunctionF := flag.String("fdistr", "const", "функция распределения чисел по горутинам")
	processFunctionF := flag.String("fproc", "mult2", "функция обработки чисел")
	// обработка введённых флагов для использования их в программе
	flag.Parse()

	// идентификаторы соответствующих функций на основе считанных из флага значений
	df := map[string]int{"const": 0, "hyp": 1}[*distrFunctionF]
	pf := map[string]int{"mult2": 0, "pow2": 1}[*processFunctionF]

	// объявление функции распределения на основе индекса
	distrFunction := []lab1.DistributionFunc{
		// функция для равномерного распределения
		func(num int) float64 {
			return 1
		},
		// функция для распределения по гиперболе
		func(num int) float64 {
			return math.Pow(1/float64(num+1), 4.5)
		},
	}[df]

	// объявление функции обработки чисел на основе считанного из флага значения
	processFunction := []lab1.ProcessingFunc{
		// функция для умножения чисел среза на 2
		func(nums []int) {
			for i := range nums {
				nums[i] *= 2
			}
			nums = nil
		},
		// функция для возведения чисел среза в квадрат
		func(nums []int) {
			for i, num := range nums {
				nums[i] = int(math.Pow(float64(num), 2))
			}
			nums = nil
		},
	}[pf]

	// запуск обработки чисел и вывод информации в консоль
	fmt.Println("--------------------------------------------------")
	fmt.Printf("\nN = %d\nM = %d\nProcess func № %d\nDistribution func № %d\n\n", *N, *M, pf+1, df+1)
	err := lab1.RunNumberProcessing(*N, *M, "testFiles", processFunction, distrFunction)
	if err != nil {
		panic(err)
	}
	fmt.Println("\n--------------------------------------------------")
}
