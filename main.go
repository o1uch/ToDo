package main

import (
	"log"
	"net/http"

	"github.com/o1uch/go_final_project/internal/api"
	"github.com/o1uch/go_final_project/internal/config"
	"github.com/o1uch/go_final_project/internal/db"
	"github.com/o1uch/go_final_project/internal/service"
	"github.com/o1uch/go_final_project/internal/store"
	_ "modernc.org/sqlite"
)

func main() {
	dbPath, err := config.GetDBPath()
	if err != nil {
		log.Fatalf("cannot get db path: %v", err)
	}

	sqlDB, err := db.Init(dbPath)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer sqlDB.Close()

	sStore := store.NewSchedulerStore(sqlDB)
	svc := service.NewService(&sStore)

	port, err := config.GetPort()
	if err != nil {
		log.Printf("port warning: %v, using default", err)
	}

	api.Init(svc)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	log.Printf("Server starting on %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
