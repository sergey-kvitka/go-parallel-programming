import datetime
import os
import subprocess as process
import time

import numpy as np

# единицы измерения времени
TIME_UNITS = ['нс', 'мкс', 'мс', 'c']

# Путь к рабочей директории
WORK_DIR = '..\\'
# Go-файл в рабочей директории, который будет запущен
EXEC_FILE = '.'
# путь к файлу для записи результата, который будет создан
RESULT_FILE = f'./result{datetime.datetime.now().strftime("%I-%M-%S%p_%B_%d_%Y")}.txt'
# путь к файлу для записи ошибок исполнения программы, который будет создан
ERROR_FILE = f'./error{datetime.datetime.now().strftime("%I-%M-%S%p_%B_%d_%Y")}.txt'
# количество повторов (задаёт количество итераций на внешнем цикле)
REPEATS = 5


def go_run_command(n, m, distr_func, proc_func):
    """
go_run_command возвращает аргументы команды для командной строке на основе аргументов для запуска обработки чисел
    """
    return ['go', 'run', EXEC_FILE, '--N', str(n), '--M', str(m), '--fdistr', distr_func, '--fproc', proc_func]


def ns_to_str(ns):
    """
ns_to_str принимает наносекунды и возвращает строку с читаемой записью времени

(например, 22937124 -> "22.937 мс")
    """
    cur_value = ns
    unit = None
    for unit in TIME_UNITS:
        if cur_value < 1000:
            break
        cur_value /= 1000
    return f'{cur_value: .2f} {unit}'


def write_to_file(filepath, text, flush=False):
    """
write_to_file записывает текст в файл по указанному местоположению.
Если flush=True, буфер файла будет очищен и вносимые в содержимое файла изменения станут видны
    """
    with open(filepath, 'a') as f:
        f.write(text)
        if flush:
            f.flush()
            os.fsync(f.fileno())


# возможные значения N (количество чисел для обработки)
n_values = [10, 100, 1_000, 10_000, 100_000, 1_000_000, 10_000_000, 100_000_000, 1_000_000_000]
# возможные значения M (количество горутин, в которых будет идти обработка чисел)
m_values = [1, 2, 3, 4, 5, 10, 20, 30, 100]
# возможные функции распределения чисел по горутинам
distr_values = ['const', 'hyp']
# возможные функции обработки чисел
proc_values = ['mult2', 'pow2']
# команда для терминала (значение будет присваиваться на каждой итерации)
command = None

# ожидаемый размер вывода запускаемой в терминале команды (количество строк)
output_size = [14, 15]

# префикс строки, содержащей информацию о времени обработки чисел
proc_res_prefix = 'Обработка чисел завершена за '

# словарь для записи результатов замеров для каждого случая
total_res = {}

# Цикл для перебора всех возможных ситуаций `REPEATS` раз
for i in range(REPEATS):
    for M in m_values:
        for proc in proc_values:
            for distr in distr_values:
                for N in n_values:
                    # исключение ситуаций, когда горутин больше, чем чисел
                    if N < M:
                        continue

                    # вывод в консоль информации о том, что итерация началась
                    print(f'\ncase [{N} {M} {proc} {distr}] has started')

                    # инициализация значения в словаре по заданному ключу в случае его отсутствия
                    if f'{N} {M} {proc} {distr}' not in total_res:
                        total_res[f'{N} {M} {proc} {distr}'] = []

                    # создание команды для запуска обработки чисел
                    command = go_run_command(N, M, distr, proc)
                    # запуск команды и получение результата запуска
                    result = process.run(command, capture_output=True, text=True, cwd=WORK_DIR, encoding='utf-8')

                    # проверка на наличие ошибки выполнения команды и запись информации в файл
                    if result.stderr != '':
                        write_to_file(ERROR_FILE,
                                      f'in case (N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}):')
                        write_to_file(ERROR_FILE,
                                      '\nstdout:\n\t' + str(result.stdout.strip()).replace('\n', '\n\t'))
                        write_to_file(ERROR_FILE,
                                      '\nstderr:\n\t' + str(result.stderr.strip()).replace('\n', '\n\t') + '\n\n',
                                      flush=True)
                        continue

                    # получение вывода программы и его разделение на отдельные строки
                    output = str(result.stdout.strip()).split('\n')
                    # проверка длины вывода и запись информации в файл в случае некорректной длины
                    if len(output) not in output_size:
                        write_to_file(ERROR_FILE,
                                      f'in case (N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}):')
                        write_to_file(ERROR_FILE,
                                      '\n\t' + str(result.stdout.strip()).replace('\n', '\n\t'))
                        write_to_file(ERROR_FILE,
                                      f'\nwrong amount of non-blank lines ({len(output)} not in {output_size})\n\n',
                                      flush=True)
                        continue

                    # получение строки с информацией о времени обработки чисел
                    proc_res = output[-3].strip()
                    # проверка полученной строки на наличие нужного префикса
                    # и запись информации в файл в случае его отсутствия
                    if not proc_res.startswith(proc_res_prefix):
                        write_to_file(ERROR_FILE,
                                      f'in case (N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}):')
                        write_to_file(ERROR_FILE,
                                      '\n\t' + str(result.stdout.strip()).replace('\n', '\n\t'))
                        write_to_file(ERROR_FILE, f'\nline "{proc_res}" does not start with "{proc_res_prefix}"\n\n',
                                      flush=True)
                        continue

                    # извлечение из полученной строки числа и единицы измерения времени (время обработки чисел)
                    proc_res = proc_res[len(proc_res_prefix):].strip().split()
                    # проверка на корректность извлечения из строки данных и запись информации в файл в противном случае
                    if len(proc_res) != 2:
                        write_to_file(ERROR_FILE,
                                      f'in case (N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}):')
                        write_to_file(ERROR_FILE,
                                      '\n\t' + str(result.stdout.strip()).replace('\n', '\n\t'))
                        write_to_file(ERROR_FILE,
                                      f'\nwrong process result format ({proc_res}) from line "{output[-3].strip()}"'
                                      f'\n\n',
                                      flush=True)
                        continue

                    # объявление переменной для записи времени обработки чисел в наносекундах
                    res_ns = 0
                    try:
                        # попытка конвертации строки в число и получения индекса элемента массива единиц измерения
                        # времени для получения нужной кратности (в итоге — перевод времени обработки файла в нс)
                        res_ns = float(proc_res[0]) * 1000 ** TIME_UNITS.index(proc_res[1])
                    # если перевод был выполнен с ошибкой, записываем соответствующую информацию в файл
                    except ValueError:
                        write_to_file(ERROR_FILE,
                                      f'\nin case (N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}):')
                        write_to_file(ERROR_FILE,
                                      '\n\t' + str(result.stdout.strip()).replace('\n', '\n\t'))
                        write_to_file(ERROR_FILE,
                                      f'\nwrong number format ({proc_res[0]}) from line "{output[-3].strip()}"\n\n',
                                      flush=True)
                        continue

                    # после прохождения всех проверок, записываем успешно извлечённые данные в файл
                    write_to_file(RESULT_FILE,
                                  f'N = {N}, M = {M}, proc_func = {proc}, distr_func = {distr}, res (ns) = {res_ns}\n',
                                  flush=True)

                    # запись полученного результата в словарь
                    total_res[f'{N} {M} {proc} {distr}'].append(res_ns)

                    # вывод в консоль информации о том, что итерация успешно завершилась
                    print(f'case [{N} {M} {proc} {distr}] completed successfully')

                # небольшой перерыв программы после каждого цикла по n_values
                print('\n\nbreak 10 sec...')
                time.sleep(10)
                print('let\'s continue\n\n')

write_to_file(RESULT_FILE, '\n\n')

# запись подытогов в файл
for key, value in total_res.items():
    parse_key = key.split()
    arr = np.array(value)
    if len(arr) == 0:
        continue
    write_to_file(RESULT_FILE,
                  f'N = {parse_key[0]}, M = {parse_key[1]}, proc_func = {parse_key[2]}, distr_func = {parse_key[3]}:\n')
    write_to_file(RESULT_FILE, f'\ttotal evaluations: {len(arr)}\n\t'
                               f'min: {np.min(arr)}\n\t'
                               f'max: {np.max(arr)}\n\t'
                               f'avg.: {np.average(arr)}\n\t'
                               f'median: {np.median(arr)}\n\t'
                               f'resulting values: {arr}\n\n')
