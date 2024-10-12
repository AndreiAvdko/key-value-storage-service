package main

import (
	"bufio"
	"fmt"
	"os"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error
	ReadEvents() (<-chan Event, <-chan error)
	Run()
}

type FileTransactionLogger struct {
	events       chan<- Event // Канал только для записи; для передачи событий
	errors       <-chan error // Канвл только для чтения; для приема ошибок
	lastSequence uint64       // Последний использованный порядковый номер
	file         *os.File     // Месторасположение файла журнала
}

type Event struct {
	Sequence  uint64    // Уникальный порядковый номер записи
	EventType EventType // Выполненное дейстие
	Key       string    //Ключ, затронутый этой транзакцией
	Value     string    // Значение для транзакции PUT
}

type EventType byte

const (
	_                     = iota // игнорируем нулевое значение
	EventDelete EventType = iota
	EventPut
)

func (l *FileTransactionLogger) WritePut(key, value string) {
	// Некоторая логика
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaxtion log file: %w", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16) // Создать канал событий
	l.events = events

	errors := make(chan error, 1) // Создать канал ошибок
	l.errors = errors

	go func() {
		for e := range events { // Извлечь следующее событие Event

			l.lastSequence++ // Увеличить порядковый номер

			_, err := fmt.Fprintf( // Записать событие в журнал
				l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}

		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file) // Создать сканнер для чтения файла
	outEvent := make(chan Event)        // Небуферизированный канал событий
	outError := make(chan error, 1)     // Буферизированный канал ошибок

	go func() {
		var e Event

		defer close(outEvent) // Закрыть каналы
		defer close(outError) // по завершении сопрограммы

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Scanf(line, "%d\t%d\t%s\t%s", &e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			// Проверка целостности!
			// Порядковые номера последовательно увеличиваются?

			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")

				return
			}

			l.lastSequence = e.Sequence // Запомнить последний использованный
			// порядковый номер

			outEvent <- e // Отправить событие along
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("ransaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}
