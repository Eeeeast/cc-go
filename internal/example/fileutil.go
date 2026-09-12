package examplelib

import (
	"bufio"
	"os"
	"sort"
)

// SortFileLines сортирует строки в указанном файле.
//
// Файл перезаписывается на месте, операция не атомарна:
// ошибка ввода-вывода при записи приведёт к потере исходных данных.
func SortFileLines(path string) (err error) {
	lines, err := readLines(path)
	if err != nil {
		return err
	}

	sort.Strings(lines)

	// Перезаписываем файл с нуля.
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	// Пробрасываем ошибку Close наружу, только если ничего другого
	// уже не упало: Close может вернуть отложенную ошибку записи на диск.
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()

	w := bufio.NewWriter(f)
	for i, line := range lines {
		if i > 0 {
			if _, err := w.WriteString("\n"); err != nil {
				return err
			}
		}
		if _, err := w.WriteString(line); err != nil {
			return err
		}
	}
	return w.Flush()
}

// readLines читает текстовый файл в кодировке UTF-8 и разбивает его на строки.
// Завершающий перевод строки не создаёт лишнюю пустую строку.
func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Дескриптор открыт только на чтение, ошибка закрытия файла не важна.
	defer func() { _ = f.Close() }()

	var lines []string
	scanner := bufio.NewScanner(f)
	// Разрешаем строки длиннее стандартных 64 КиБ.
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}