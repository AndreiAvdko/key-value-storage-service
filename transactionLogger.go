package main

import (
	"fmt"
	"os"
)

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
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
