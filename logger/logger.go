// logger/logger.go
package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
)

var (
	logger     *logrus.Logger
	loggerOnce sync.Once
)

// GetLogger возвращает глобальный логгер с ленивой инициализацией
func GetLogger() *logrus.Logger {
	loggerOnce.Do(func() {
		logger = logrus.New()
		logger.SetLevel(logrus.DebugLevel)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})

		// Настройка вывода лога в файл
		file, err := os.OpenFile("app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			logger.Error("Ошибка открытия файла лога", err)
			return
		}
		logger.SetOutput(file)
	})
	return logger
}
