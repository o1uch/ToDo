package config

import (
	"fmt"
	"os"
	"strconv"
)

func GetPort() (string, error) {
	defPort := 7540
	currPort := 0
	envPort := os.Getenv("TODO_PORT")

	if envPort == "" {
		currPort = defPort
		return fmt.Sprintf(":%d", currPort), nil
	}

	currPort, err := strconv.Atoi(envPort)

	if err != nil {
		currPort = defPort
		return fmt.Sprintf(":%d", currPort), fmt.Errorf("ошибка конвертации значения порта из переменной TODO_PORT: %w\nБудет использован путь по умолчанию", err)
	}
	return fmt.Sprintf(":%d", currPort), nil
}
