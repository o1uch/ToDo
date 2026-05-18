package postgres

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var (
	ConnStr = "postgresql://scheduler_usr:pg4todo@localhost:5433/scheduler?sslmode=disable" // для простоты разработки и отладки пока строка
)

func Init(connStr string) (*sql.DB, error) {

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("error opening the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error checking the availability of the database: %w", err)
	}

	if err := createSchedulerSchema(db); err != nil {
		return nil, fmt.Errorf("error initializing the scheduler schema: %w", err)
	}

	return db, nil
}

func createSchedulerSchema(db *sql.DB) error {

	createSchema := `create table if not exists scheduler (
id SERIAL primary key,
date varchar(8) not null default '',
title varchar(256) not null default '',
comment text not null default '', 
repeat varchar(128) not null default '');

create index if not exists date_index on scheduler (date); `

	_, err := db.Exec(createSchema)

	if err != nil {
		return err
	}

	return nil
}
