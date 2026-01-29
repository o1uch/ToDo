package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetDBPath() (string, error) {
	var currDBPath, envDBPath string
	defaultDBPath := "scheduler.db"
	dbName := "scheduler.db"

	if envDBPath = os.Getenv("TODO_DBFILE"); envDBPath == "" {
		return defaultDBPath, nil
	}

	envDBPath, err := filepath.Abs(envDBPath)

	if err != nil {
		return defaultDBPath, fmt.Errorf("ошибка получения пути к БД из переменной окружения:%w\nБудет использован путь по умолчанию", err)
	}

	file, err := os.Stat(envDBPath)

	if err != nil {
		if os.IsNotExist(err) {
			return envDBPath, nil
		}
		return defaultDBPath, fmt.Errorf("ошибка получения пути к БД из переменной окружения:%w\nБудет использован путь по умолчанию", err)
	}

	if file.IsDir() {
		currDBPath = filepath.Join(envDBPath, dbName)
		return currDBPath, nil
	}

	return envDBPath, nil
}
