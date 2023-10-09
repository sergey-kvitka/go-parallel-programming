package main

import (
	"fmt"
	"go-parallel-programming/lab0"
	"go-parallel-programming/lab1"
	"go-parallel-programming/lesson-tasks/lessons/linklist"
	"go-parallel-programming/lesson-tasks/lessons/tasks"
	"math"
	"time"
)

func main() {

	//lab1Test()
	a := time.Now()

	fmt.Println("init...")
	arr := [1e9]int{}
	fmt.Println(time.Since(a))
	a = time.Now()
	fmt.Println("set values...")
	for i, _ := range arr {
		arr[i] = i
	}
	fmt.Println(time.Since(a))
	a = time.Now()
	fmt.Println("x2...")
	for i, _ := range arr {
		arr[i] *= 2
	}
	fmt.Println(time.Since(a))
	a = time.Now()
	fmt.Println("x2...")
	for i, _ := range arr {
		arr[i] *= 2
	}
	fmt.Println(time.Since(a))
	a = time.Now()
	fmt.Println("x3...")
	for i, _ := range arr {
		arr[i] *= 3
	}
	fmt.Println(time.Since(a))
	a = time.Now()
}

func lab1Test() {

	var (
		multiplier, exponentiation lab1.ProcessingFunc
		linear, exponential        lab1.DistributionFunc
	)

	multiplier = func(nums []int) {
		for i, _ := range nums {
			nums[i] *= 2
		}
		nums = nil
	}
	exponentiation = func(nums []int) {
		for i, num := range nums {
			nums[i] = int(math.Pow(float64(num), 2))
		}
		nums = nil
	}

	linear = func(num int) float64 {
		return 1
	}
	exponential = func(num int) float64 {
		return math.Exp(float64(num))
	}

	var (
		fileDir = "testFiles"
		//numAmounts            = []int{1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9}
		numAmounts            = []int{1e9}
		routineAmounts        = []int{1, 2, 3, 4, 5, 10, 20, 30, 100}
		processFunctions      = []*lab1.ProcessingFunc{&multiplier, &exponentiation}
		distributionFunctions = []*lab1.DistributionFunc{&linear, &exponential}
		err                   error
	)

	for df, distrFunc := range distributionFunctions {
		for pf, processFunc := range processFunctions {
			for _, N := range numAmounts {
				for _, M := range routineAmounts {
					if M > N {
						continue
					}
					fmt.Println("--------------------------------------------------")
					fmt.Printf("\nN = %d\nM = %d\nProcess func № %d\nDistribution func № %d\n\n", N, M, pf+1, df+1)

					err = lab1.RunNumberProcessing(N, M, fileDir, *processFunc, *distrFunc)
					if err != nil {
						panic(err)
					}

					fmt.Println("\n--------------------------------------------------")
				}
			}
		}
	}
}

func lab0Test() {

	//brackets := "()[(){{[]}}]"
	//fmt.Println(lab0.IsValidParenthesesString(brackets))

	triangleRows := 5
	fmt.Println(lab0.PascalTriangle(triangleRows))

	//islandMap := [][]int{
	//	{1, 1, 1, 0},
	//	{1, 0, 0, 0},
	//	{1, 0, 0, 0},
	//	{1, 1, 0, 0},
	//}
	//fmt.Println(lab0.IslandPerimeter(islandMap))
	//
	//intToRoman := 3888
	//fmt.Println(lab0.IntToRoman(intToRoman))
	//
	//nums := []int{1231, 34, 64, 123, 5, 658, 58, 5, 67, 32, 2, 1}
	//fmt.Println(lab0.ArrangeToLargestNumber(nums))

}

func linklistTest() {
	linkedList := linklist.List{}

	linkedList.Add(0)
	linkedList.Add(1)
	linkedList.Add(2)
	linkedList.Add(3)
	linkedList.Add(4)
	linkedList.Add(5)
	linkedList.Add(6)
	linkedList.Add(7)
	linkedList.Add(8)
	linkedList.Add(9)

	var err error
	fmt.Println(linkedList.ToSlice())
	err = linkedList.Set(77, 7)
	fmt.Println(linkedList.ToSlice())
	err = linkedList.Set(44, 4)
	fmt.Println(linkedList.ToSlice())
	err = linkedList.Insert(100, 9)
	fmt.Println(linkedList.ToSlice())
	err = linkedList.Insert(333, 7)
	fmt.Println(linkedList.ToSlice())
	deleted1, err := linkedList.Delete(6)
	fmt.Println(deleted1)
	fmt.Println(linkedList.ToSlice())
	deleted2, err := linkedList.Delete(2)
	fmt.Println(deleted2)
	fmt.Println(linkedList.ToSlice())
	linkedList.Add(150)
	fmt.Println(linkedList.ToSlice())
	linkedList.Add(250)
	length := linkedList.Length()
	fmt.Println(length)
	fmt.Println(linkedList.ToSlice())
	if err != nil {
		panic(err)
	}
}

func sumArrTest() {

	var (
		start time.Time
		n     = 1000000000
		sum   int
	)

	nums := make([]int, n)

	fmt.Println("Writing numbers...")
	for i := 0; i < n; i++ {
		nums[i] = i
	}
	fmt.Print("Done\n\n")

	start = time.Now()
	for _, num := range nums {
		sum += num
	}
	fmt.Printf("Sum (unparallel): %d\n", sum)
	fmt.Printf("Time elapsed (unparallel): %d microseconds\n", time.Since(start).Microseconds())

	start = time.Now()
	sum = tasks.ParallelSliceSumChan(nums)
	fmt.Printf("Sum (parallel): %d\n", sum)
	fmt.Printf("Time elapsed (parallel): %d microseconds\n", time.Since(start).Microseconds())

}
