package logx

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	once      sync.Once
	logger    *zap.Logger
	errLogger error
)

// TO-DO: добавить возможность настроить имя файла лога!
func Get(fileName string) (*zap.Logger, error) {
	once.Do(
		func() {
			// конфиги енкодеров
			consolCfg := zap.NewDevelopmentEncoderConfig()
			prodCfg := zap.NewProductionEncoderConfig()

			// настройка write-синкеров
			file, errLogger := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if errLogger != nil {
				return
			}
			// defer file.Close()
			jsonSync := zapcore.AddSync(file)

			consolSync := zapcore.AddSync(os.Stdout)

			// кодировщики для записи логов
			consolEnc := zapcore.NewConsoleEncoder(consolCfg)
			jsonEnc := zapcore.NewJSONEncoder(prodCfg)

			// настройка ядра
			core := zapcore.NewTee(
				zapcore.NewCore(consolEnc, consolSync, zap.NewAtomicLevelAt(zapcore.DebugLevel)),
				zapcore.NewCore(jsonEnc, jsonSync, zap.NewAtomicLevelAt(zapcore.InfoLevel)),
			)
			logger = zap.New(core)
		})
	return logger, errLogger
}
