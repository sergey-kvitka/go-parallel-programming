package lab1

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	FilePathTemplate = "%s/numbers%d.txt"
)

// ! шаблоны текста ошибок
const (
	_GenNumbersErrTemplate = "неверный формат входных данных (n = %d, expected n ≥ 1, chunkSize = %d, expected chunkSize ≥ 1, (writer == nil) is %v, expected false)"
	_DistributeErrTemplate = "неверный формат входных данных (totalNumber = %d, elementsNumber = %d, expected totalNumber ≥ elementsNumber & totalNumber ≥ 1, (distribution == nil) is %v, expected false)"
	_DistrValueErrTemplate = "функция распределения не должна принимать неположительные значения при неотрицательном аргументе (distribution(%v) = %v)"
	_RNPErrTemplate        = "неверный формат входных данных (N = %d, M = %d, expected 10 ≤ N ≤ 1,000,000,000 & 1 ≤ M ≤ 1000 & M ≤ N, (processingFunction == nil) is %v, expected false, (numberDistribution == nil) is %v, expected false)"
)

// WriterFunc является consumer'ом для среза чисел и может вернуть ошибку;
// используется для обработки записи срезов чисел (например, в определённую структуру или в файл)
type WriterFunc func(nums []int) error

// DistributionFunc принимает целое число и возвращает число с плавающей точкой;
// используется при вычислении распределения чисел по элементам среза
type DistributionFunc func(num int) float64

// ProcessingFunc является consumer'ом для среза целых чисел;
// используется для выполнения операций над числами при их обработке в отдельных горутинах
type ProcessingFunc func(nums []int)

// goroutineStartFunc используется для запуска обработки среза чисел в отдельной горутине;
// помимо среза чисел nums принимает *sync.WaitGroup starter, после ожидания которого будет
// запущена обработка чисел, и *sync.WaitGroup wg, у которого будет вызван соответствующий
// метод, обозначающий полное завершение обработки чисел
type goroutineStartFunc func(nums []int, starter *sync.WaitGroup, wg *sync.WaitGroup)

// GenerateNumbers генерирует натуральные числа от 1 до n включительно и последовательно вызывает
// функцию writer для каждой созданной последовательности чисел (далее 'кусок') размером chunkSize.
//
// Например, при аргументах n = 10 и chunkSize = 4 функция writer будет вызвана с
// передачей в неё последовательностей [1, 2, 3, 4], [5, 6, 7, 8] и [9, 10].
func GenerateNumbers(n int, chunkSize int, writer WriterFunc) error {
	// проверка входных параметров на корректность
	if n < 1 || chunkSize < 1 || writer == nil {
		return fmt.Errorf(_GenNumbersErrTemplate, n, chunkSize, writer == nil)
	}
	var (
		left             = n   // количество оставшихся чисел, не разделённых по кускам
		currentChunk     []int // текущий кусок
		currentChunkSize int   // количество чисел, которое выносится в отдельный кусок в данный момент
		currentNumber    = 1   // текущее число
		err              error // ошибка
	)
	// цикл до тех пор, пока чисел не останется
	for left != 0 {
		// определение размера куска для записи (берётся меньшее из
		// переданного в функцию chunkSize и количества оставшихся чисел)
		if chunkSize < left {
			currentChunkSize = chunkSize
		} else {
			currentChunkSize = left
		}
		// обнуление текущего куска (размер задаётся на основе определённого ранее размера куска)
		currentChunk = make([]int, currentChunkSize)
		// запись нужных чисел в текущий кусок (с инкрементацией текущего числа)
		for i := range currentChunk {
			currentChunk[i] = currentNumber
			currentNumber++
		}
		// уменьшаем количество оставшихся чисел на размер выделенного куска
		left -= currentChunkSize
		// вызов writer с передачей в него выделенного куска и перехват возможной ошибки
		err = writer(currentChunk)
		if err != nil {
			return err
		}
	}
	return nil // всё прошло без ошибок, возвращается nil
}

// GetFileWriter возвращает функцию WriterFunc для записи последовательностей
//  чисел в файл. Каждое число записывается в файл на отдельной строке.
func GetFileWriter(file *os.File) WriterFunc {
	return func(nums []int) error { // объявление возвращаемой функции
		// создание буфера для дальнейшей записи в файл
		var buffer bytes.Buffer
		// в цикле идёт последовательная запись в буфер числа и переноса строки
		for _, num := range nums {
			buffer.WriteString(strconv.Itoa(num))
			buffer.WriteString("\n")
		}
		// запись в переданный файл байт из сформированного буфера с последующей проверкой на ошибку
		_, err := file.Write(buffer.Bytes())
		if err != nil {
			return err
		}
		return nil // всё прошло без ошибок, возвращается nil
	}
}

// Distribute распределяет N чисел (N = totalNumber) на M частей (M = elementsNumber) по заданной функции
// распределения DistributionFunc. На каждую часть гарантированно придётся хотя бы по 1 числу.
func Distribute(totalNumber int, elementsNumber int, distribution DistributionFunc) ([]int, error) {
	// проверка входных параметров на корректность
	if totalNumber < elementsNumber || totalNumber < 1 || distribution == nil {
		return nil, fmt.Errorf(_DistributeErrTemplate, totalNumber, elementsNumber, distribution == nil)
	}
	var (
		evalValues  = make([]float64, elementsNumber) // просчитанные значения функции распределения
		result      = make([]int, elementsNumber)     // срез для записи результата распределения
		value       float64                           // текущее просчитанное значение
		ratio       float64                           // соотношение суммы просчитанных значений и общего кол-ва чисел
		evalSum     float64                           // сумма просчитанных значений
		decimalPart float64                           // текущая десятичная часть от просчитанного значения
		decimalSum  float64                           // сумма накопленных десятичных частей
	)
	// цикл по элементам evalValues для просчёта значений
	for i := range evalValues {
		// вычисление текущего значения (на основе индекса элемента) и проверка значения на корректность
		value = distribution(i)
		if value <= 0 {
			return nil, fmt.Errorf(_DistrValueErrTemplate, i, value)
		}
		// сохранение вычисленного значения и увеличение суммы всех просчитанных значений
		evalValues[i] = value
		evalSum += value
	}
	// вычисление соотношения общего количества чисел и суммы просчитанных
	// значений, при этом из общего количества вычитается количество частей,
	// чтобы гарантировать ненулевые значения (≥ 1) для каждой части
	ratio = float64(totalNumber-elementsNumber) / evalSum

	// цикл по элементам result для просчёта итоговых значений
	for i := range result {
		// расчёт значения на основе перемножения соответствующего просчитанного
		// через функцию распределения элемента на полученное соотношение ratio,
		// а также прибавление единицы для гарантии ненулевого значения
		value = 1 + evalValues[i]*ratio
		// в результирующий срез записывается целая часть от полученного значения
		result[i] = int(value)
		// вычисление оставшейся дробной части и увеличение суммы дробных частей
		decimalPart = value - float64(result[i])
		decimalSum += decimalPart
		// если дробная часть стала больше 1, то увеличиваем текущий элемент
		// результирующего среза на 1 и также уменьшаем сумму десятичных частей
		if decimalSum > 1 {
			result[i]++
			decimalSum--
		}
	}
	// * если округление остатка от суммы десятичных частей даёт 1, увеличиваем последний элемент результата на 1;
	// данное действие совершается по той причине, что при работе с числами с плавающей точкой имеется
	// небольшая погрешность при вычислении значений (например, 0.2 + 0.3 = 0.5000000001); по этой причине
	// идёт дополнительная проверка суммы: если сумма равна 0.999, то она не пройдёт последнюю проверку в цикле,
	// но при этом мы явно можем сказать, что данное число меньше 1 лишь из-за погрешности вычислений, и потому
	// мы можем просто прибавить 1 к последнему элементу; если погрешность была в другую сторону и мы получили,
	// например, 0.06, то при округлении мы получим 0, а значит дополнительных действий не будет совершено
	if math.Round(decimalSum) == 1 {
		result[elementsNumber-1]++
	}
	return result, nil // возвращаем результирующий срез
}

// RunNumberProcessing запускает многопоточную обработку чисел, которая включает в себя следующие шаги:
//
// 1. Создание файла, содержащего N чисел, в директории fileDir, или получение уже существующего файла;
//
// 2. Распределение N чисел по M горутинам на основе функции распределения numberDistribution;
//
// 3. Считывание ранее полученного файла, формирование горутин на основе распределения чисел
// и функции для их обработки processingFunction;
//
// 4. Запуск сформированных горутин.
//
// Информация о выполнении шагов 1, 3 и 4 выводится в консоль
// (в том числе и информация о времени выполнения данных шагов).
func RunNumberProcessing(
	// количество натуральных чисел, которое будет сгенерировано и записано в файл (10 ≤ N ≤ 1,000,000,000)
	N int,
	// количество горутин, в которых будут обрабатываться сгенерированные числа (1 ≤ M ≤ 1000, M ≤ N)
	M int,
	// путь к директории, в которой находится или будет создан файл с числами
	fileDir string,
	// функция обработки чисел (в отдельной горутине)
	processingFunction ProcessingFunc,
	// функция распределения чисел по горутинам (необязательно равномерная)
	numberDistribution DistributionFunc,
) error {
	// проверка входных параметров на корректность
	if N < 10 || M < 1 || N > 1e9 || M > 1e3 || M > N || processingFunction == nil || numberDistribution == nil {
		return fmt.Errorf(_RNPErrTemplate, N, M, processingFunction == nil, numberDistribution == nil)
	}

	var (
		filepath = fmt.Sprintf(FilePathTemplate, fileDir, N) // путь к файлу с числами
		fileSize int64                                       // размер файла (в байтах)
		err      error                                       // ошибка
	)

	// создания файла и запись в него N чисел (если файл с таким расположением существует, новый файл создан не будет)
	fileSize, err = writeNumbersToFile(filepath, N)
	if err != nil {
		return err
	}

	// открытие созданного файла для дальнейшего чтения
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	// отложенная функция закрытия файла
	defer func() {
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()

	// распределение N чисел по M горутинам по функции numberDistribution
	distributedNumbers, err := Distribute(N, M, numberDistribution)
	if err != nil {
		return err
	}

	// считывание файла с числами для формирования M горутин и запуск сформированных горутин
	err = processNumbersFromFile(
		file,
		fileSize,
		distributedNumbers,
		func(nums []int, starter *sync.WaitGroup, wg *sync.WaitGroup) {
			starter.Wait()
			defer wg.Done()
			processingFunction(nums)
		},
	)
	return err // возврат значения err (будет равно nil при корректном завершении работы предыдущей операции)
}

// writeNumbersToFile создаёт (при необходимости) файл с расположением
// filepath и записывает в него N (N = numbersAmount) натуральных чисел
// от 1 до N включительно. Возвращает размер полученного файла в байтах
func writeNumbersToFile(filepath string, numbersAmount int) (int64, error) {
	var (
		file     *os.File    // файл для записи чисел
		fileInfo os.FileInfo // информация о файле
		err      error       // ошибка
	)
	// отложенная функция закрытия файла (с проверкой на nil)
	defer func() {
		if file == nil {
			return
		}
		if err = file.Close(); err != nil {
			panic(err)
		}
	}()
	// попытка получения информации о файле с расположением filepath
	fileInfo, err = os.Stat(filepath)
	// проверка случая, когда файла с расположением filepath не существует;
	// при таком сценарии будут выполнены создание файла и запись в него чисел
	if errors.Is(err, os.ErrNotExist) {
		// создание файла с расположением filepath (с проверкой на ошибку)
		if file, err = os.Create(filepath); err != nil {
			return 0, err
		}
		fmt.Printf("Генерация файла и запись в него чисел (количество чисел: %d)...\n", numbersAmount)
		// в случае, если количество чисел (далее N) большое (> 1000) запись в файл будет произведена по частям;
		// размер одной части равен ⌈2 × √(N)⌉, т.е. удвоенному значению квадратного корня из N, округлённому вверх;
		// если количество чисел – 1000 или меньше, числа в файл будут записаны единой строкой
		var chunkSize int
		if numbersAmount > 1000 {
			chunkSize = int(math.Ceil(math.Sqrt(float64(numbersAmount))) * 2)
		} else {
			chunkSize = numbersAmount
		}
		startWF := time.Now() // ! засекаем время записи чисел в файл
		// запись чисел в созданный файл по частям
		err = GenerateNumbers(numbersAmount, chunkSize, GetFileWriter(file))
		elapsedWF := time.Since(startWF) // ! засекаем время записи чисел в файл
		if err != nil {
			return 0, err
		}
		fmt.Printf("Файл успешно сгенерирован за %s (расположение файла: \"%s\")\n", elapsedWF, filepath)
		// получение информации о созданном файле (с проверкой на ошибку)
		if fileInfo, err = os.Stat(filepath); err != nil {
			return 0, err
		}
	} else
	// если же при получении информации о файле была получена другая ошибка
	// (не свидетельствующая о том, что файла не существует), возвращаем данную ошибку
	if err != nil {
		return 0, err
	} else
	// если ошибки при получении информации о файле не возникло (что свидетельствует
	// о том, что он существует), просто выводим информацию в консоль
	{
		fmt.Printf(
			"Файл с нужным количеством чисел (%d) уже существует, новый файл не будет сгенерирован (расположение файла: \"%s\")\n",
			numbersAmount, filepath,
		)
	}

	return fileInfo.Size(), nil // возвращаем размер полученного/созданного файла
}

// processNumbersFromFile выполняет операции чтения файла с числами, распределения
// считанных чисел по горутинам и одновременного запуска данных горутин (при этом
// засекается время их работы). Информация о работе функции выводится в консоль.
func processNumbersFromFile(
	// файл с числами
	file *os.File,
	// размер файла (в байтах)
	fileSize int64,
	// распределение чисел по горутинам (каждый элемент соответствует количеству чисел, которое будет выделено горутине)
	distributedNumbers []int,
	// функция для контролируемого через sync.WaitGroup запуска горутин
	goroutineStarter goroutineStartFunc,
) error {
	// в случае, если размер файла в байтах (далее S) большое (> 1,000,000) чтение файла будет произведено по частям;
	// размер одной части равен ⌈2 × √(S)⌉, т.е. удвоенному значению квадратного корня из S, округлённому вверх;
	// если размер файла – 1,000,000 байт или меньше, файл будет считан полностью за 1 операцию
	var chunkSize int64
	if fileSize > 1e6 {
		chunkSize = int64(math.Ceil(math.Sqrt(float64(fileSize))) * 2)
	} else {
		chunkSize = fileSize
	}

	var (
		reader            = bufio.NewReader(file)   // объект для чтения файла
		byteBuffer        = make([]byte, chunkSize) // буфер считанных байтов
		prevBytes         []byte                    // необработанные байты, считанные на предыдущей итерации
		numbersFromFile   []int                     // текущие считанные из файла числа
		strNumsBuffer     []string                  // буфер считанных из файла чисел в строковом виде
		numbersFromStr    []int                     // конвертированные в числа строковые представления считанных чисел
		nBytes            int                       // количество считанных байтов
		newlineIndex      int                       // индекс последнего символа новой строки в считанном куске файла
		goroutinesAmount  = len(distributedNumbers) // количество горутин, которое будет выделено
		distributionIndex = 0                       // индекс числа в distributedNumbers (для выделения нужного кол-ва чисел)
		starter           = &sync.WaitGroup{}       // WaitGroup для одновременного запуска горутин
		waitGroup         = &sync.WaitGroup{}       // WaitGroup для ожидания выполнения всех горутин
		err               error                     // ошибка
	)

	// взведение счётчика WaitGroup запуска горутин (все горутины будут ждать обнуления счётчика)
	starter.Add(1)
	// взведение счётчика WaitGroup ожидания завершения горутин (все горутины будут
	// уменьшать счётчик на 1, и нулевое значение будет означать завершение их работы)
	waitGroup.Add(goroutinesAmount)

	fmt.Printf("Чтение файла...\n")
	startFR := time.Now() // ! засекаем время чтения файла
	// запуск цикла для последовательного чтения файла по частям; выход из цикла будет произведён тогда, когда при
	// очередном чтении куска файла будет считано 0 байтов (что будет означать, что reader дошёл до конца файла)
	for {
		// чтение файла с записью считанных байтов в буфер
		nBytes, err = io.ReadFull(reader, byteBuffer)
		// возвращаем ошибку, если чтение файла дало ошибку, не означающую конец файла (io.EOF или io.ErrUnexpectedEOF)
		if err != nil && !(errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)) {
			return err
		}
		// если было считано 0 файлов, то, при наличии ранее считанных из файла и необработанных чисел, запускаем
		// ещё одну горутину и увеличиваем индекс числа в distributedNumbers, после чего выходим из цикла
		if nBytes == 0 {
			if len(numbersFromFile) != 0 {
				go goroutineStarter(numbersFromFile, starter, waitGroup)
				distributionIndex++
			}
			break
		}

		// получаем индекс последнего символа новой строки в считанной с файла последовательности байтов
		newlineIndex = strings.LastIndex(string(byteBuffer[:nBytes]), "\n")
		// если такого символа не нашлось, то добавляем считанные байты к срезу ранее считанных и необработанных байтов
		if newlineIndex == -1 {
			prevBytes = append(prevBytes, byteBuffer[:nBytes]...)
			continue // переход к следующему чтению файла
		}

		// конкатенируем строку из предыдущих байтов с удалёнными из начала символами конца строки и строку из только
		// что считанных из файла байтов, после чего разделяем полученную строку на срез, используя в качестве
		// разделителя символ конца строки; в итоге получаем срез чисел в строковом представлении
		strNumsBuffer = strings.Split(strings.TrimLeft(string(prevBytes), "\n")+string(byteBuffer[:newlineIndex]), "\n")
		// в цикле переводим числа из строкового представления в целочисленное, записывая их при этом в срез
		numbersFromStr = make([]int, len(strNumsBuffer))
		for i, strNum := range strNumsBuffer {
			numbersFromStr[i], err = strconv.Atoi(strNum)
			if err != nil { // проверка на возникновение ошибки при конвертации строки в число
				return err
			}
		}
		// в срез чисел из файла добавляем только что конвертированные строки
		numbersFromFile = append(numbersFromFile, numbersFromStr...)

		// выполняем цикл до тех пор, пока считанных на данный момент из файла чисел хватает для передачи их в
		// следующую горутину в соответствии со значением в срезе распределения чисел
		for distributionIndex < goroutinesAmount && len(numbersFromFile) >= distributedNumbers[distributionIndex] {
			// запуск горутины с нужным количеством чисел из ранее считанных из файла чисел
			go goroutineStarter(numbersFromFile[:distributedNumbers[distributionIndex]], starter, waitGroup)
			// уменьшение среза считанных из файла чисел после передачи части чисел в горутину
			numbersFromFile = numbersFromFile[distributedNumbers[distributionIndex]:]
			// увеличение индекса числа в distributedNumbers
			distributionIndex++
		}

		// перезапись среза предыдущих байтов; данный срез будет состоять из байтов между
		// последним символом конца строки и концом считанной из файла последовательности байтов
		prevBytes = make([]byte, nBytes-newlineIndex-1)
		copy(prevBytes, byteBuffer[newlineIndex+1:nBytes])
	}

	// если после работы цикла индекс числа в distributedNumbers не равен количеству горутин,
	// возвращаем ошибку, сигнализирующую о некорректном завершении работы цикла
	if distributionIndex != goroutinesAmount {
		return errors.New("ошибка распределения потоков: нужное количество потоков не было выделено")
	}
	elapsedFR := time.Since(startFR) // ! засекаем время чтения файла
	fmt.Printf("Файл считан за %s\n", elapsedFR)

	fmt.Printf("Обработка чисел (количество потоков: %d)...\n", goroutinesAmount)
	startNP := time.Now() // ! засекаем время выполнения всех горутин
	// данной операцией одновременно запускаем все ранее созданные горутины
	starter.Done()
	// ожидаем выполнения всех запущенных горутин
	waitGroup.Wait()
	elapsedNP := time.Since(startNP) // ! засекаем время выполнения всех горутин
	fmt.Printf("Обработка чисел завершена за %s\n", nsToStr(elapsedNP.Nanoseconds()))
	return nil // всё прошло без ошибок, возвращается nil
}

// nsToStr принимает наносекунды и возвращает строку с читаемой записью времени
// (например, 22937124 -> "22.937 мс")
func nsToStr(ns int64) string {
	var unit string
	curValue := float64(ns)
	for _, unit = range timeUnits {
		if curValue < 1000 {
			break
		}
		curValue /= 1000
	}
	return fmt.Sprintf("%.3f %s", curValue, unit)
}

// timeUnits хранит единицы измерения времени в порядке возрастания
var timeUnits = []string{"нс", "мкс", "мс", "c"}
