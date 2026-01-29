package main

import (
	"fmt"
	"net/http"

	"github.com/o1uch/go_final_project/internal/config"
	"github.com/o1uch/go_final_project/internal/db"
	_ "modernc.org/sqlite"
)

func main() {
	// реализация шага 2. Создание БД.

	dbPath, err := config.GetDBPath()

	if err != nil {
		fmt.Println(err)
		return
	}

	db, err := db.Init(dbPath)

	if err != nil {
		fmt.Println(err)
		return
	}

	if err := db.Ping(); err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("connection successfuly")
	}

	// шаг 1
	webDir := "./web"
	port, _ := config.GetPort()

	http.Handle("/", http.FileServer(http.Dir(webDir)))

	err = http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

}
