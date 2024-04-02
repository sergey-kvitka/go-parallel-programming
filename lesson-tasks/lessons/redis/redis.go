package redis

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// timeFormat - формат даты и времени, используемый для записи и чтения даты при работе с бэкапами
const timeFormat string = "02.01.2006 15:04:05.000 MST"

// symbols - объект, в котором хранятся токены, используемые для сохранения файла в бэкап,
// а также алгоритмы экранирования и деэкранирвоания (в текущей реализации это не кастомизируется)
var symbols = escaping{
	record:    "?",
	delimiter: "&",
	end:       "!",
	escape:    url.QueryEscape,
	unescape:  url.QueryUnescape,
}

// escaping является типом выше описанного файла
type escaping struct {
	record    string
	delimiter string
	end       string
	escape    func(string) string
	unescape  func(string) (string, error)
}

// BackupOptions используется для передачи необходимой информации для сохранения и восстановления данных;
// передается при создании экземпляра redis
type BackupOptions struct {
	BackupFolder   string
	BackupFileName string
	Interval       time.Duration
	ResultChan     chan AsyncResult
}

// AsyncResult - результат, возвращаемый по каналу BackupOptions.ResultChan при восстановлении и сохранении данных
// (необходимо настраивать вывод значений из канала)
type AsyncResult struct {
	Process string
	Status  string
	Error   error
}

// KeyType - возможные типы ключей
type KeyType interface {
	int | float64 | string
}

// ValueType - возможные типы значений
type ValueType interface {
	int | float64 | bool | string
}

type Instance[K KeyType, V ValueType] struct {
	data map[K]V
	mtx  *sync.RWMutex

	timeouts timeoutInfo[K]
	interval time.Duration

	backupOptions *BackupOptions
}

// NewInstance - основной конструктор
func NewInstance[K KeyType, V ValueType](
	interval time.Duration,
	backupOptions *BackupOptions,
) Instance[K, V] {
	instance := Instance[K, V]{
		data: make(map[K]V),
		mtx:  &sync.RWMutex{},

		timeouts: newTimeout[K](),
		interval: interval,
	}

	if backupOptions != nil {
		instance.backupOptions = backupOptions
		instance.restoreBackup()
		instance.startBackupCycle()
	}

	return instance
}

// NewInstance1S - конструктор для создания экземпляра без бэкапов и ежесекундной проверкой данных для удаления
func NewInstance1S[K KeyType, V ValueType]() Instance[K, V] {
	return NewInstance[K, V](time.Second, nil)
}

func (r *Instance[K, V]) Get(key K) (V, bool) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	value, found := r.data[key]
	return value, found
}

// Set записывает значение по ключу. Если такой ключ уже существует, то значение будет перезаписано.
// Если на ключ наложен timeout для его удаления, вызов данного метода отменит удаление и удалит информацию о timeout.
func (r *Instance[K, V]) Set(key K, value V) {
	r.mtx.Lock()
	r.timeouts.mtx.Lock()

	defer r.mtx.Unlock()
	defer r.timeouts.mtx.Unlock()

	r.data[key] = value
	delete(r.timeouts.data, key)
}

// SetWithTimeout записывает значение по ключу и время, когда этот ключ будет удален, на основе переданного expiration.
// Если такой ключ уже существует, значение будет перезаписано и время удаление будет записано/перезаписано.
func (r *Instance[K, V]) SetWithTimeout(key K, value V, expiration time.Duration) {
	if expiration <= 0 {
		return
	}

	r.mtx.Lock()
	r.timeouts.mtx.Lock()

	defer r.mtx.Unlock()
	defer r.timeouts.mtx.Unlock()

	r.data[key] = value
	r.timeouts.data[key] = time.Now().Add(expiration)
	if len(r.timeouts.data) == 1 {
		r.startTimeoutCheck()
	}
}

// startBackupCycle запускает бесконечный цикл с интервалом для периодического сохранения данных в файл.
// Запускается 1 раз при создании экземпляра в том случае, если необходимая информация была передана при создании.
func (r *Instance[K, V]) startBackupCycle() {
	go func() {
		for range time.Tick(r.backupOptions.Interval) {
			r.backup()
		}
	}()
}

// startTimeoutCheck запускает цикл с интервалом для проверки наличия записей,
// которые по времени уже необходимо удалить, и удаляет найденные записи
func (r *Instance[K, V]) startTimeoutCheck() {
	go func() {

		timeoutData := r.timeouts.data
		timeoutMtx := r.timeouts.mtx

		for now := range time.Tick(r.interval) {
			timeoutMtx.Lock()

			for k, v := range timeoutData {
				if !now.Before(v) {
					r.mtx.Lock()
					delete(timeoutData, k)
					delete(r.data, k)
					r.mtx.Unlock()
				}
			}

			if len(timeoutData) == 0 {
				timeoutMtx.Unlock()
				return
			}
			timeoutMtx.Unlock()
		}
	}()
}

// getBackupFile используется для получения файла бэкапа.
// Будет создан пустой файл при отсутствии такового в директории backupFileName.
func getBackupFile(backupFolder string, backupFileName string, clearFile bool) (*os.File, error) {
	folder, err := os.Stat(backupFolder)
	if err != nil {
		return nil, err
	}
	if !folder.IsDir() {
		return nil, fmt.Errorf("file with path \"%v\" is not a directory", backupFolder)
	}

	filePath := backupFolder + "\\" + backupFileName

	var file *os.File
	if clearFile {
		file, err = os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755) // пытаемся получить файл
	} else {
		file, err = os.Open(filePath) // пытаемся получить файл
	}

	if err != nil { // если ошибка, создаем файл
		if file, err = os.Create(filePath); err != nil { // создание файла (+ err check)
			return nil, err
		}
	}
	return file, nil
}

// backup выполняет сохранение данных в файл, расположенный в директории по переданному пути backupFolder.
// Процесс запускается в отдельном потоке и не блокирует вызываемые операции записи и чтения данных.
//
// Информация об успешном или неудачном выполнении
// сохранения данных будет передана через канал BackupOptions.ResultChan.
func (r *Instance[K, V]) backup() {
	processName := "backup"
	success := map[bool]string{
		true:  "success",
		false: "fail",
	}

	go func(backupFolder string, backupFileName string, result chan AsyncResult) {
		file, err := getBackupFile(backupFolder, backupFileName, true)
		if err != nil {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error:   err,
			}
			return
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		var k_ K
		var v_ V
		_, err = file.WriteString(fmt.Sprintf( // запись в файл строки с информацией о типах ключа и значения
			"%v %v\n",
			reflect.TypeOf(k_).String(),
			reflect.TypeOf(v_).String(),
		))
		if err != nil {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error:   err,
			}
			return
		}

		var keys []K
		r.mtx.Lock() // считываем ключи мапы для дальнейшей работы с ней
		keys = make([]K, len(r.data))
		i := 0
		for k := range r.data {
			keys[i] = k
			i++
		}
		r.mtx.Unlock()

		for _, k := range keys {
			var timeStr string
			r.timeouts.mtx.Lock() // * thread safe timeout access
			if timeout, ok := r.timeouts.data[k]; ok {
				timeStr = timeout.Format(timeFormat)
			}
			r.timeouts.mtx.Unlock() // *.

			v, ok := r.Get(k) // thread safe value access
			if !ok {
				continue
			}
			_, err = file.WriteString(fmt.Sprint(
				symbols.record,
				url.QueryEscape(fmt.Sprint(any(k))), symbols.delimiter,
				url.QueryEscape(fmt.Sprint(any(v))), symbols.delimiter,
				url.QueryEscape(timeStr),
			))
			if err != nil {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}
		}

		if _, err = file.WriteString(symbols.end); err != nil {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error:   err,
			}
			return
		}

		result <- AsyncResult{
			Process: processName,
			Status:  success[true],
			Error:   nil,
		}

	}(r.backupOptions.BackupFolder, r.backupOptions.BackupFileName, r.backupOptions.ResultChan)
}

// restoreBackup выполняет восстановление данных из файла, расположенного в директории по переданному пути backupFolder.
// Процесс запускается в отдельном потоке и не блокирует вызываемые операции записи и чтения данных.
//
// Информация об успешном или неудачном выполнении восстановления
// данных будет передана через канал BackupOptions.ResultChan.
func (r *Instance[K, V]) restoreBackup() {
	processName := "restoreBackup"
	success := map[bool]string{
		true:  "success",
		false: "fail",
	}

	go func(backupFolder string, backupFileName string, result chan AsyncResult) {

		file, err := getBackupFile(backupFolder, backupFileName, false)
		if err != nil {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error:   err,
			}
			return
		}
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

		reader := bufio.NewReader(file)

		bytes, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error:   err,
			}
			return
		}

		if err == io.EOF && len(bytes) == 0 {
			result <- AsyncResult{
				Process: processName,
				Status:  "empty",
				Error:   nil,
			}
			return
		}

		typesStr := strings.Trim(string(bytes), " \n")
		types := strings.Split(typesStr, " ")
		if len(types) != 2 {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error: fmt.Errorf("first line of backup file must contain key and value types info "+
					"(e.g. \"string int\", \"int float64\"), got \"%s\"", typesStr),
			}
			return
		}
		var (
			k_              K
			v_              V
			keyType         = reflect.TypeOf(k_).String()
			valueType       = reflect.TypeOf(v_).String()
			backupKeyType   = strings.Trim(types[0], " ")
			backupValueType = strings.Trim(types[1], " ")
		)

		if keyType != backupKeyType {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error: fmt.Errorf("wrong key type info from backup file (expected \"%s\", got \"%s\")",
					keyType, backupKeyType),
			}
			return
		}
		if valueType != backupValueType {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error: fmt.Errorf("wrong value type info from backup file (expected \"%s\", got \"%s\")",
					valueType, backupValueType),
			}
			return
		}

		recordByte := []byte(symbols.record)[0]
		endByte := []byte(symbols.end)[0]

		nextByte, err := reader.ReadByte()
		if err != nil {
			return
		}
		if nextByte != recordByte {
			result <- AsyncResult{
				Process: processName,
				Status:  success[false],
				Error: fmt.Errorf("wrong record beginning symbol (expected '%v', got '%v')",
					recordByte, nextByte),
			}
			return
		}

		notEOF := true
		for notEOF {
			bytes, err = reader.ReadBytes(recordByte)
			if err != nil && err != io.EOF {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}
			if bytes[len(bytes)-1] == endByte || err == io.EOF {
				notEOF = false
			}

			data := strings.Trim(string(bytes), symbols.record+symbols.end)
			dataArr := strings.Split(data, symbols.delimiter)
			if len(dataArr) != 3 {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error: fmt.Errorf("unable to parse backup record "+
						"(can't extract 3 tokens from \"%v\" with separator \"%v\")", data, symbols.delimiter),
				}
				return
			}

			var keyStr, valueStr, dateStr string
			keyStr, err = symbols.unescape(dataArr[0])
			if err != nil {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}
			valueStr, err = symbols.unescape(dataArr[1])
			if err != nil {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}
			dateStr, err = symbols.unescape(dataArr[2])
			if err != nil {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}

			var (
				k    K
				v    V
				date time.Time
			)
			k, v, err = convert[K, V](keyStr, valueStr)
			if err != nil {
				result <- AsyncResult{
					Process: processName,
					Status:  success[false],
					Error:   err,
				}
				return
			}

			if dateStr != "" {
				date, err = time.Parse(timeFormat, dateStr)
				if err != nil {
					result <- AsyncResult{
						Process: processName,
						Status:  success[false],
						Error:   err,
					}
					return
				}
				r.SetWithTimeout(k, v, date.Sub(time.Now()))
			} else {
				r.Set(k, v)
			}
		}

		result <- AsyncResult{
			Process: processName,
			Status:  success[true],
			Error:   nil,
		}

	}(r.backupOptions.BackupFolder, r.backupOptions.BackupFileName, r.backupOptions.ResultChan)
}

// convert конвертирует строки с ключом и значением в типы K и V соответственно
func convert[K KeyType, V ValueType](keyStr string, valueStr string) (K, V, error) {
	var (
		key   K
		value V
	)

	switch any(key).(type) {
	case string:
		key = any(keyStr).(K)
		break
	case int:
		number, err := strconv.Atoi(keyStr)
		if err != nil {
			return key, value, err
		}
		key = any(number).(K)
		break
	case float64:
		number, err := strconv.ParseFloat(keyStr, 64)
		if err != nil {
			return key, value, err
		}
		key = any(number).(K)
	}

	switch any(value).(type) {
	case string:
		value = any(valueStr).(V)
		break
	case int:
		number, err := strconv.Atoi(valueStr)
		if err != nil {
			return key, value, err
		}
		value = any(number).(V)
		break
	case float64:
		number, err := strconv.ParseFloat(valueStr, 64)
		if err != nil {
			return key, value, err
		}
		value = any(number).(V)
		break
	case bool:
		b, err := strconv.ParseBool(valueStr)
		if err != nil {
			return key, value, err
		}
		value = any(b).(V)
	}

	return key, value, nil
}
