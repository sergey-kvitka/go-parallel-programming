package lab0

import (
	"sort"
	"strconv"
	"strings"
)

// ! Лабораторная работа №0

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// * Задача №1
// ? Сложность: easy
// ! Название:  Valid Parentheses
/* * Условия:
* Given a string s containing just the characters '(', ')', '{', '}', '[' and ']', determine if the input string is valid.
* An input string is valid if:
*    Open brackets must be closed by the same type of brackets.
*    Open brackets must be closed in the correct order.
*    Every close bracket has a corresponding open bracket of the same type.
? https://leetcode.com/problems/valid-parentheses/description
*/

// IsValidParenthesesString проверяет строку, состоящую из скобок, на корректность открытия и закрытия
// всех скобок. Возвращает true, если последовательность скобок корректна, и false, если
// последовательность скобок некорректна или строка является пустой или содержит иные символы
func IsValidParenthesesString(str string) bool {
	return isValid(str)
}

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// * Задача №2
// ? Сложность: easy
// ! Название:  Pascal's Triangle
/* * Условия:
* Given an integer numRows, return the first numRows of Pascal's triangle.
? https://leetcode.com/problems/pascals-triangle/description
*/

// PascalTriangle возвращает срез, содержащий numRows срезов чисел,
// каждый из которых представляет собой один ряд из треугольника Паскаля
func PascalTriangle(numRows int) [][]int {
	return generate(numRows)
}

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// * Задача №3
// ? Сложность: easy
// ! Название:  Island Perimeter
/* * Условия:
* You are given row x col grid representing a map where grid[i][j] = 1 represents land and grid[i][j] = 0 represents water.
* Grid cells are connected horizontally/vertically (not diagonally). The grid is completely surrounded by water,
* and there is exactly one island (i.e., one or more connected land cells).
*
* The island doesn't have "lakes", meaning the water inside isn't connected to the water around the island.
* One cell is a square with side length 1. The grid is rectangular, width and height don't exceed 100. Determine the perimeter of the island.
? https://leetcode.com/problems/island-perimeter/description
*/

// IslandPerimeter возвращает площадь острова, записанного в виде сетки из единиц и нулей, где 1 - это суша, а 0 - вода
func IslandPerimeter(grid [][]int) int {
	return islandPerimeter(grid)
}

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// * Задача №4
// ? Сложность: medium
// ! Название:  Integer to Roman
/* * Условия:
* Roman numerals are represented by seven different symbols: I, V, X, L, C, D and M.
* For example, 2 is written as II in Roman numeral, just two one's added together. 12 is written as XII, which is
* simply X + II. The number 27 is written as XXVII, which is XX + V + II.
*
* Roman numerals are usually written largest to smallest from left to right. However, the numeral for four is not IIII.
* Instead, the number four is written as IV. Because the one is before the five we subtract it making four.
* The same principle applies to the number nine, which is written as IX. There are six instances where subtraction is used:
*    I can be placed before V (5) and X (10) to make 4 and 9.
*    X can be placed before L (50) and C (100) to make 40 and 90.
*    C can be placed before D (500) and M (1000) to make 400 and 900.
*
* Given an integer, convert it to a roman numeral.
? https://leetcode.com/problems/integer-to-roman/description
*/

// IntToRoman переводит целое число num в римское число (например, 294 -> "CCXCIV")
//
// Если число num не входит в диапазон [1, 3999], будет возвращена пустая строка
func IntToRoman(num int) string {
	return intToRoman(num)
}

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// * Задача №5
// ? Сложность: medium
// ! Название:  Largest Number
/* * Условия:
* Given a list of non-negative integers nums, arrange them such that they form the largest number and return it.
*
* Since the result may be very large, so you need to return a string instead of an integer.
? https://leetcode.com/problems/largest-number/description
*/

// ArrangeToLargestNumber возвращает число в виде строки, являющееся результатом конкатенации чисел из переданного среза
// таким образом, что в результате получается наибольшее возможное число (например, {30, 3, 35} -> "35330")
func ArrangeToLargestNumber(nums []int) string {
	return largestNumber(nums)
}

// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————
// * ————————————————————————————————— Далее идёт реализация методов из условий задач —————————————————————————————————
// * ——————————————————————————————————————————————————————————————————————————————————————————————————————————————————

// bracketInfo является структурой для информации о скобках
type bracketInfo struct {
	pairValue int32
	opened    bool
}

// bracketMap содержит пары из символов скобок и информации о них (ссылка на объект bracketInfo).
// Информация включает в себя идентификатор пары скобок (например, у обеих скобок '[' и ']' это значение будет '[')
// и то, является ли эта скобка открывающейся (true или false).
var bracketMap = map[int32]*bracketInfo{
	'{': {pairValue: '{', opened: true},
	'}': {pairValue: '{', opened: false},
	'[': {pairValue: '[', opened: true},
	']': {pairValue: '[', opened: false},
	'(': {pairValue: '(', opened: true},
	')': {pairValue: '(', opened: false},
}

// isValid проверяет строку, состоящую из скобок, на корректность открытия и закрытия всех скобок.
// Возвращает true, если последовательность скобок корректна, и false, если последовательность
// скобок некорректна или строка является пустой или содержит иные символы
func isValid(s string) bool {
	// пустая строка или строка нечётной длины не может быть валидной
	if len(s) == 0 || len(s)%2 == 1 {
		return false
	}

	// стек текущих открытых скобок
	var bracketsStack []int32

	// цикл по символам из пришедшей строки
	for _, bracket := range s {
		var (
			stackLength        = len(bracketsStack)  // текущая длина стека
			currentBracketInfo = bracketMap[bracket] // информация (из мапы) о текущей скобке из строки
			stackBracketInfo   *bracketInfo          // информация (из мапы) о последней скобке из стека
		)
		// если стек пустой, сохраняем ссылку на значение из мапы в stackBracketInfo
		if stackLength != 0 {
			stackBracketInfo = bracketMap[bracketsStack[stackLength-1]]
		}
		// если в мапе не нашлось информации о текущем символе из пришедшей строки, возвращаем false
		if currentBracketInfo == nil {
			return false
		}
		// если текущая скобка открывающаяся, добавляем её в стек и переходим на новую итерацию
		if currentBracketInfo.opened {
			bracketsStack = append(bracketsStack, bracket)
			continue
		}
		// если нет информации о последней скобке в стеке (т.е. стек пустой)
		// или скобки из строки и из стека не являются парой, возвращаем false
		if stackBracketInfo == nil || currentBracketInfo.pairValue != stackBracketInfo.pairValue {
			return false
		}
		// если все проверки пройдены, то скобки из стека и из строки являются единой парой
		// из открывающейся и закрывающейся скобок, поэтому убираем одну скобку из стека
		bracketsStack = bracketsStack[:stackLength-1]
	}
	// если стек полностью очищен (пустой), значит все скобки закрыты, строка
	// является корректной и метод возвратит true, в противном случае возвратится false
	return len(bracketsStack) == 0
}

// generate возвращает срез, содержащий numRows срезов чисел,
// каждый из которых представляет собой один ряд из треугольника Паскаля
func generate(numRows int) [][]int {
	var (
		result  [][]int // результирующий срез срезов чисел
		prevRow *[]int  // ссылка на предыдущий ряд треугольника
	)
	// цикл по натуральным числам до numRows включительно
	for i := 1; i <= numRows; i++ {
		// для текущего ряда создаём пустой срез длиной i и записываем в его первый и последний элементы значение 1
		row := make([]int, i)
		row[0] = 1
		row[i-1] = 1
		// если ряд состоит из 3 и более элементов, то во все его элементы (их текущий индекс в цикле - k), кроме
		// первого и последнего, записываем значение, равное сумме k-го и (k - 1)-го элементов предыдущего ряда
		if i > 2 {
			for k := 1; k < i-1; k++ {
				row[k] = (*prevRow)[k-1] + (*prevRow)[k]
			}
		}
		// переопределяем ссылку на предыдущий ряд и добавляем ряд в результирующий срез срезов чисел
		prevRow = &row
		result = append(result, row)
	}
	return result
}

// islandPerimeter возвращает площадь острова, записанного в виде сетки из единиц и нулей, где 1 - это суша, а 0 - вода
func islandPerimeter(grid [][]int) int {
	var (
		rows      = len(grid) // количество рядов в сетке
		columns   int         // количество столбцов в сетке
		perimeter int         // итоговый периметр
	)
	// стандартный цикл в цикле по сетке grid
	for i := 0; i < rows; i++ {
		columns = len(grid[i])
		for j := 0; j < columns; j++ {
			// если текущая ячейка - вода, то переходим на следующую итерацию
			if grid[i][j] == 0 {
				continue
			}
			// увеличиваем периметр, если ячейка находится в самом верхнем ряду или в ячейке на 1 ряд выше вода
			if i == 0 || grid[i-1][j] == 0 {
				perimeter++
			}
			// увеличиваем периметр, если ячейка находится в самом левом столбце или в ячейке на 1 столбец левее вода
			if j == 0 || grid[i][j-1] == 0 {
				perimeter++
			}
			// увеличиваем периметр, если ячейка находится в самом нижнем ряду или в ячейке на 1 ряд ниже вода
			if i == rows-1 || grid[i+1][j] == 0 {
				perimeter++
			}
			// увеличиваем периметр, если ячейка находится в самом правом столбце или в ячейке на 1 столбец правее вода
			if j == columns-1 || grid[i][j+1] == 0 {
				perimeter++
			}
		}
	}
	return perimeter
}

// romanNumberPowerInfo является структурой для информации о разряде римских чисел
type romanNumberPowerInfo struct {
	one  int32
	five int32
}

// romanNumbersInfo содержит пары из номеров разряда и информации о них (ссылка на объект romanNumberPowerInfo).
// Информация включает в себя символ для единицы (1) данного разряда и символ для пятёрки (5) данного разряда.
var romanNumbersInfo = map[int]romanNumberPowerInfo{
	1: {one: 'I', five: 'V'},
	2: {one: 'X', five: 'L'},
	3: {one: 'C', five: 'D'},
	4: {one: 'M'},
}

// intToRoman переводит целое число num в римское число (например, 294 -> "CCXCIV")
//
// Если число num не входит в диапазон [1, 3999], будет возвращена пустая строка
func intToRoman(num int) string {
	if num < 1 || num > 3999 {
		return ""
	}
	var (
		tenPower           = 1     // номер текущего разряд
		current            = num   // текущее значение конвертируемого числа
		digit              int     // текущий разряд
		result             []int32 // срез для записи результата
		currentDigitResult []int32 // промежуточный результат перевода отдельного разряда
	)
	// в конце каждой итерации увеличиваем номер текущего разряда
	for ; ; tenPower++ {
		// текущий разряд из остатка от деления текущего числа на 10
		digit = current % 10
		// обнуляем промежуточный результат
		currentDigitResult = []int32{}
		// особым образом обрабатываем случай, когда текущий разряд можно
		// записать в римском числе с вычитанием (как, например, IV или IX);
		// выполняется, когда digit = 4 или digit = 9
		if (digit+1)%5 == 0 {
			// римское число, которое будет записано перед вычитанием
			var numberToSubtract int32
			// если текущий разряд равен 4, то перед вычитанием запишется
			// число, представляющее собой 5 на текущем номере разряда (V, L и т.д.)
			if digit == 4 {
				numberToSubtract = romanNumbersInfo[tenPower].five
			} else
			// если текущий разряд равен 9, то перед вычитанием запишется
			// число, представляющее собой 1 на следующем номере разряда (X, C и т.д.)
			if digit == 9 {
				numberToSubtract = romanNumbersInfo[tenPower+1].one
			}
			// в промежуточный результат записываем 2 символа -
			// число, представляющее собой 1 на текущем номере разряда,
			// и число перед ним на основе значения текущего разряда
			// (например: ['I', 'X'] (IX - 9), ['X', 'L'] ((XL - 40)))
			currentDigitResult = append(currentDigitResult, romanNumbersInfo[tenPower].one, numberToSubtract)
		} else
		// обработка остальных случаев (когда текущий разряд (digit) не равен 4 или 9)
		{
			// если текущий разряд больше или равен 5, добавляем в промежуточный
			// результат число, представляющее собой 5 на текущем разряде
			if digit >= 5 {
				currentDigitResult = append(currentDigitResult, romanNumbersInfo[tenPower].five)
			}
			// добавляем в промежуточный результат N чисел, представляющих собой 1 на текущем разряде;
			// N равно остатку от деления текущего разряда на 5
			for i := 0; i < digit%5; i++ {
				currentDigitResult = append(currentDigitResult, romanNumbersInfo[tenPower].one)
			}
		}
		// добавляем в срез для записи результата промежуточный результат
		result = append(currentDigitResult, result...)
		// делим текущее число на 10 и выходим из цикла, если результате целочисленного деления получился 0
		current = current / 10
		if current == 0 {
			break
		}
	}
	// возвращаем строку, состоящую из символов, сохранённых в срез для записи результата
	return string(result)
}

// largestNumber возвращает число в виде строки, являющееся результатом конкатенации чисел из переданного среза
// таким образом, что в результате получается наибольшее возможное число (например, {30, 3, 35} -> "35330")
func largestNumber(nums []int) string {
	// срез для сохранения чисел в строковом представлении
	strNums := make([]string, 0, len(nums))
	// цикл по срезу чисел для сохранения их строковых представлений
	for _, num := range nums {
		strNums = append(strNums, strconv.Itoa(num))
	}
	// сортировка полученных чисел по заданному алгоритму;
	// функция, передаваемая в метод сортировки, принимает индексы i и j сравниваемых элементов и возвращает true,
	// если алгоритм сортировки подразумевает, что элемент на i-й позиции после сортировки должен стоять перед
	// j-м элементом; если 2 элемента в любом порядке возвращают false, то они считаются равными
	sort.Slice(strNums, func(i, j int) bool {
		// создаём 2 переменные, являющиеся конкатенацией соответствующих элементов среза,
		// чтобы найти более подходящий порядок элементов
		a := strNums[i] + strNums[j]
		b := strNums[j] + strNums[i]
		length := len(a)

		// итерация по элементам данных строк
		for k := 0; k < length; k++ {
			// находим разницу кода соответствующих символов (код 0 меньше кода 9)
			diff := int(a[k]) - int(b[k])
			// если полученная разница ненулевая, то возвращаем true, если она больше 0, или false, если она меньше 0
			if diff != 0 {
				return diff > 0
			}
		}
		// если в итоге все символы равны и выхода из метода сравнения не произошло, просто возвращаем false
		return false
	})
	// результатом будет являться соединённые элементы среза чисел с удалёнными спереди нулями на случай,
	// если на вход функции было передано несколько нулей, и необходимо отбросить незначащие нули
	result := strings.TrimLeft(strings.Join(strNums, ""), "0")
	// в итоге если результат оказался пустым, просто возвращаем "0", а иначе возвращаем сам результат
	if result == "" {
		return "0"
	}
	return result
}
