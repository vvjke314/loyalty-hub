package logx

import (
	"io"
	"os"
	"testing"

	"encoding/json"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type log struct {
	Level     string  `json:"level"`
	Timestamp float64 `json:"ts"`
	Message   string  `json:"msg"`
}

const (
	InfoLevel  = "INFO"
	DebugLevel = "DEBUG"
	ErrorLevel = "ERROR"
)

func inputWithLevel(logger *zap.Logger, message, level string) {
	switch level {
	case InfoLevel:
		logger.Info(message)
	case ErrorLevel:
		logger.Error(message)
	case DebugLevel:
		logger.Debug(message)
	}
}

// добавить тест на инициализацию логгера и проверку реализации singleton'a

// TO-DO: использовать tempdir для тестов
func TestLoggerInput(t *testing.T) {
	// настройка среды для тестов
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal(err)
	}

	fileName := "test.log"

	// запуск создания тестов
	logger, err := Get(fileName)
	if err != nil {
		t.Fatal(err)
	}

	// подготовка тест-кейсов
	tests := []struct {
		name  string
		level string
		input string
	}{
		{
			name:  "success_case",
			level: InfoLevel,
			input: "hello world",
		},
	}

	// запускаем сабтесты
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// вписываем данные
			inputWithLevel(logger, tt.input, tt.level)
			// высвобождение ресурсов и запись в мсто назначения
			_ = logger.Sync()

			// проверяем данные
			file, err := os.OpenFile(fileName, os.O_RDONLY, 0644)
			if err != nil {
				t.Fatalf("can't open file %v %s", err, fileName)
			}
			defer file.Close()
			raw, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			// если дебаг лвл, то не должно отображаться в json'e
			if tt.level == DebugLevel && len(raw) > 2 {
				t.Error("expected zero input in file")
				return
			}
			var l log
			if err := json.Unmarshal(raw, &l); err != nil || l.Message != tt.input {
				t.Errorf("values are not equal %v. expected %s got %s", err, tt.input, l.Message)
			}

			// очищаем файл (truncate до 0 байт)
			if err := os.Truncate(fileName, 0); err != nil {
				t.Fatalf("failed to truncate file: %v", err)
			}
		})
	}
}
