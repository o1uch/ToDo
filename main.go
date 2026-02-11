package main

import (
	"fmt"
	"net/http"

	"github.com/o1uch/go_final_project/internal/api"
	"github.com/o1uch/go_final_project/internal/config"
	"github.com/o1uch/go_final_project/internal/db"
	"github.com/o1uch/go_final_project/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
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

	store := store.NewSchedulerStore(db)

	webDir := "./web"
	port, _ := config.GetPort()

	api.Init(&store)
	http.Handle("/", http.FileServer(http.Dir(webDir)))

	err = http.ListenAndServe(port, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
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



			-- NextDate tests

		now := time.Date(2024, 01, 26, 00, 0, 0, 0, time.UTC)

		dstart := "20240116"

		repeatValue := "m -1,18"

		nextDate, err := repeat.NextDate(now, dstart, repeatValue)

		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(nextDate)

	*/
}
