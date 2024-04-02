package redis

import (
	"fmt"
	"github.com/google/uuid"
	"os"
	"reflect"
	"sync"
	"testing"
	"time"
)

func getBackupDir(t *testing.T) string {
	wd, err := os.Getwd()
	if err != nil {
		t.Error(err)
		panic(err)
	}
	return wd + "\\backup"
}

func Timeout(exp time.Duration, t *testing.T, stop chan bool) { // на случай deadlock
	errChan := make(chan bool)

	go func() {
		time.Sleep(exp)
		errChan <- true
	}()

	go func() {
		<-stop
		errChan <- false
	}()

	go func() {
		res := <-errChan
		if res {
			err := fmt.Errorf("timeout error (too long evaluation, might be deadlock)")
			t.Error(err)
			panic(err)
		}
		<-errChan
	}()
}

// Test1 делает проверку действия Set параллельно в нескольких потоках и сверяет количество добавленных записей
func Test1(t *testing.T) {
	stop := make(chan bool)
	Timeout(10*time.Second, t, stop)

	r := NewInstance1S[int, int]()
	w := sync.WaitGroup{}
	w.Add(4)

	go func() {
		defer w.Done()
		r.Set(1, 50)
		r.Set(2, 60)
		r.Set(3, 70)
		r.Set(4, 80)
	}()
	go func() {
		defer w.Done()
		r.Set(4, 90)
		r.Set(5, 100)
		r.Set(6, 110)
		r.Set(7, 120)
	}()
	go func() {
		defer w.Done()
		r.Set(7, 130)
		r.Set(8, 140)
		r.Set(9, 150)
		r.Set(10, 160)
	}()
	go func() {
		defer w.Done()
		r.Set(10, 170)
		r.Set(11, 180)
		r.Set(12, 190)
		r.Set(13, 200)
	}()

	w.Wait()

	expectedLength := 13
	realLength := len(r.data)

	if realLength != expectedLength {
		t.Errorf("wrong records amount after multithread write (expected %v, got %v)",
			expectedLength, realLength)
	} else {
		t.Logf("records amount is correct (%v)", realLength)
	}

	stop <- true
}

// Test2 во множестве потоков заносит данные из мапы в экземпляр redis и проверяет, что каждый элемент занесён корректно
func Test2(t *testing.T) {
	stop := make(chan bool)
	Timeout(10*time.Second, t, stop)

	data := map[int]int{ // данные для проверки
		10: 10, 20: 100,
		30: 1_000, 40: 10_000,
		50: 100_000, 60: 1_000_000,
		70: 20, 80: 200,
		90: 2_000, 100: 20_000,
		110: 200_000, 120: 2_000_000,
		130: 30, 140: 300,
		150: 3_000, 160: 30_000,
		170: 300_000, 180: 3_000_000,
		190: 40, 200: 400,
	}

	length := len(data)

	var keys []int
	for k := range data {
		keys = append(keys, k)
	}

	r := NewInstance1S[int, int]()

	w := sync.WaitGroup{}
	w.Add(length / 2)

	i := 0
	for i < (length - length%2) {
		go func(i int) {
			defer w.Done()
			r.Set(keys[i], data[keys[i]])
			r.Set(keys[i+1], data[keys[i+1]])
		}(i)
		i += 2
	}
	if i == length-1 {
		r.Set(keys[i], data[keys[i]])
	}

	w.Wait()

	for k, v := range data {
		value, ok := r.Get(k)
		if !ok {
			t.Errorf("failed to save {%v: %v} (key %v not found)", k, v, k)
		} else if v != value {
			t.Errorf("failed to save {%v: %v} (expected value %v, got %v)", k, v, v, value)
		}
	}

	stop <- true
}

// Test3 в цикле заносит в redis набор значений с определённым timeout (бо́льшим на каждой итерации), а затем
// через такие же промежутки времени сверяет количество записей в redis и количество записей в массиве timeout
// (где хранится информация о времени, до которого нужно хранить определённые записи)
func Test3(t *testing.T) {
	stop := make(chan bool)
	Timeout(40*time.Second, t, stop)

	r := NewInstance1S[int, int]()

	r.Set(1_000_000, 1_000_000)
	r.Set(2_000_000, 2_000_000)
	r.Set(3_000_000, 3_000_000)

	staticAmount := len(r.data)

	r.Set(2, 22)
	r.Set(4, 44)
	r.Set(7, 77)

	currentAmount := 10

	for i := 1; i <= currentAmount; i++ {
		r.SetWithTimeout(i, i*10, time.Second*time.Duration(i))
	}

	time.Sleep(time.Second / 2)

	for currentAmount >= 0 {
		if len(r.data)-staticAmount != currentAmount {
			t.Errorf("1")
		}
		if len(r.timeouts.data) != currentAmount {
			t.Errorf("2")
		}

		time.Sleep(time.Second)
		currentAmount--
	}

	stop <- true
}

// Test4 проверяет работу функционала сохранения бэкапа с определённым периодом и
// восстановлением данных из файла бэкапа при создании нового экземпляра
func Test4(t *testing.T) {
	stop := make(chan bool)
	Timeout(20*time.Second, t, stop)

	resChan := make(chan AsyncResult)
	handleRes := func(res AsyncResult) {
		switch res.Status {
		case "success":
			t.Logf("Process \"%v\" successfully completed", res.Process)
		case "fail":
			t.Errorf("Process \"%v\" failed (%v)", res.Process, res.Error.Error())
		case "empty":
			t.Logf("Process \"%v\" has stopped: file is empty", res.Process)
		}
	}

	options := &BackupOptions{
		BackupFolder:   getBackupDir(t),
		BackupFileName: t.Name() + "_" + uuid.New().String()[:6] + ".txt",
		Interval:       time.Second * 2,
		ResultChan:     resChan,
	}

	r := NewInstance[string, int](time.Second, options)
	r.Set("a", 10)
	r.Set("b", 20)
	r.Set("c", 30)

	handleRes(<-resChan) // бэкап при создании

	handleRes(<-resChan) // бэкап через 2 секунды

	r2 := NewInstance[string, int](time.Second, options)

	handleRes(<-resChan) // бэкап при создании нового экземпляра

	if v, ok := r2.Get("a"); !ok || v != 10 {
		t.Errorf("Key \"%v\" not found", 10)
	}
	if v, ok := r2.Get("b"); !ok || v != 20 {
		t.Errorf("Key \"%v\" not found", 20)
	}
	if v, ok := r2.Get("c"); !ok || v != 30 {
		t.Errorf("Key \"%v\" not found", 30)
	}

	stop <- true
}

// Test5 проверяет корректность работы сохранения и восстановления данных
// в и из бэкапа в связке со временными данными (которые удаляются через определённое время)
func Test5(t *testing.T) {
	stop := make(chan bool)
	Timeout(50*time.Second, t, stop)

	resChan := make(chan AsyncResult)
	go func() {
		for {
			res := <-resChan
			switch res.Status {
			case "success":
				t.Logf("Process \"%v\" successfully completed", res.Process)
			case "fail":
				t.Errorf("Process \"%v\" failed (%v)", res.Process, res.Error.Error())
			case "empty":
				t.Logf("Process \"%v\" has stopped: file is empty", res.Process)
			}
		}

	}()

	backupFolder := getBackupDir(t)
	backupFileName := t.Name() + "_" + uuid.New().String()[:6] + ".txt"

	/// 0
	r1 := NewInstance[string, int](time.Second/5, &BackupOptions{
		BackupFolder:   backupFolder,
		BackupFileName: backupFileName,
		Interval:       time.Second * 5,
		ResultChan:     resChan,
	})

	time.Sleep(time.Second)
	/// ≈1 секунд от начала

	r1.Set("1", 30)
	r1.Set("2", 60)
	r1.Set("3", 90)

	time.Sleep(time.Second * 5)
	/// ≈6 секунд от начала

	r2 := NewInstance[string, int](time.Second/5, &BackupOptions{
		BackupFolder:   backupFolder,
		BackupFileName: backupFileName,
		Interval:       time.Second * 5,
		ResultChan:     resChan,
	})

	time.Sleep(time.Second)
	/// ≈7 секунд от начала

	if !reflect.DeepEqual(r1.data, r2.data) {
		t.Errorf("COMPARISON 1 (7 sec): %v != %v", r1.data, r2.data)
	} else {
		t.Logf("COMPARISON 1 (7 sec): %v = %v", r1.data, r2.data)
	}

	r2.SetWithTimeout("2", 61, time.Second/4)
	r2.SetWithTimeout("3", 91, time.Second*5)
	r2.SetWithTimeout("4", 120, time.Second*10)
	r2.SetWithTimeout("5", 150, time.Second*17+time.Second/2)

	time.Sleep(time.Second)
	/// ≈8 секунд от начала

	if _, ok := r2.Get("2"); ok {
		t.Error("Key '2': failed to delete")
	} else {
		t.Logf("Key '2' successfully deleted")
	}

	time.Sleep(time.Second * 5)
	/// ≈13 секунд от начала

	r3 := NewInstance[string, int](time.Second/5, &BackupOptions{
		BackupFolder:   backupFolder,
		BackupFileName: backupFileName,
		Interval:       time.Second * 5,
		ResultChan:     resChan,
	})

	time.Sleep(time.Second)
	/// ≈14 секунд от начала

	r3Data := map[string]int{
		"1": 30,
		"4": 120,
		"5": 150,
	}
	if !reflect.DeepEqual(r3.data, r3Data) {
		t.Errorf("COMPARISON 2 (14 sec): %v != %v", r3.data, r3Data)
	} else {
		t.Logf("COMPARISON 2 (14 sec): %v = %v", r3.data, r3Data)
	}

	time.Sleep(time.Second * 5)
	/// ≈19 секунд от начала

	r4 := NewInstance[string, int](time.Second/5, &BackupOptions{
		BackupFolder:   backupFolder,
		BackupFileName: backupFileName,
		Interval:       time.Second * 5,
		ResultChan:     resChan,
	})
	time.Sleep(time.Second * 2)
	/// ≈21 секунд от начала

	r4Data := map[string]int{
		"1": 30,
		"5": 150,
	}
	if !reflect.DeepEqual(r4.data, r4Data) {
		t.Errorf("COMPARISON 3 (21 sec): %v != %v", r4.data, r4Data)
	} else {
		t.Logf("COMPARISON 3 (21 sec): %v = %v", r4.data, r4Data)
	}

	time.Sleep(time.Second * 5)
	/// ≈26 секунд от начала

	r5 := NewInstance[string, int](time.Second/5, &BackupOptions{
		BackupFolder:   backupFolder,
		BackupFileName: backupFileName,
		Interval:       time.Second * 5,
		ResultChan:     resChan,
	})
	time.Sleep(time.Second)
	/// ≈27 секунд от начала

	r5Data := map[string]int{
		"1": 30,
	}
	if !reflect.DeepEqual(r5.data, r5Data) {
		t.Errorf("COMPARISON 5 (27 sec): %v != %v", r5.data, r5Data)
	} else {
		t.Logf("COMPARISON 5 (27 sec): %v = %v", r5.data, r5Data)
	}

	stop <- true
}
