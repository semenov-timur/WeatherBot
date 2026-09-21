// Package config отвечает за загрузку, парсинг и валидацию
// конфигурации всего приложения из переменных окружения.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config содержит все глобальные настройки, необходимые для работы приложения.
type Config struct {
	// TelegramToken — секретный токен для авторизации Telegram-бота.
	TelegramToken string
}

// Load загружает конфигурацию из переменных окружения, проводит её валидацию
// и возвращает заполненную структуру [Config].
//
// Если обязательные переменные окружения отсутствуют или содержат только пробелы,
// метод возвращает ошибку.
func Load() (Config, error) {
	token, err := requireEnv("TELEGRAM_BOT_TOKEN")
	if err != nil {
		return Config{}, err
	}
	return Config{TelegramToken: token}, nil
}

// requireEnv извлекает значение переменной окружения по ключу key.
// Возвращает ошибку, если переменная не задана или после удаления
// концевых пробелов оказалась пустой.
func requireEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("environment variable %s is not set", key)
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("environment variable %s is empty", key)
	}
	return value, nil
}
