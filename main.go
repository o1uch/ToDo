package main

import (
	"fmt"
	"time"

	"github.com/o1uch/go_final_project/internal/repeat"
	_ "modernc.org/sqlite"
)

func main() {

	// тест правила "y"

	now := time.Date(2026, 02, 04, 11, 0, 0, 0, time.UTC)

	dstart := "20260129"

	repeatValue := "w "

	nextDate, err := repeat.NextDate(now, dstart, repeatValue)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(nextDate)

	/*

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

	*/
}
