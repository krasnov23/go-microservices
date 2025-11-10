package main

import (
	"authentication/data"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib" // Регистрирует драйвер "pgx" для database/sql, если убрать sql получим: unknown driver "pgx"
)

const webPort = "80"

var counts int64

type Config struct {
	DB     *sql.DB
	Models data.Models
}

func main() {
	log.Println("Starting auhtentication service")

	conn := connectToDB()

	if conn == nil {
		log.Panic("Cant connect to postgres")
	}

	app := Config{
		DB:     conn,
		Models: data.New(conn),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()

	if err != nil {
		log.Panic(err)
	}
}

func openDB(dsn string) (*sql.DB, error) {

	// открывает объект подключения с использованием драйвера "pgx".
	// соединение ещё не устанавливается — просто создаётся структура для него.
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}

	// тестирует подключение к БД (выполняет короткий запрос).
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// В итоге connectToDB() делает 10 попыток подключения с интервалом в 2 секунды.
// Это нужно, если БД поднимается позже, например, при запуске через Docker Compose.
func connectToDB() *sql.DB {

	// получает строку подключения из переменной окружения DSN
	dsn := os.Getenv("DSN")

	for {
		connection, err := openDB(dsn)
		if err != nil {
			log.Println("Postgres not yet ready...")
			counts++
		} else {
			log.Println("Connected to database!")
			return connection
		}

		if counts > 10 {
			log.Println(err)
			return nil
		}

		log.Println("Backing off for 2 seconds...")
		time.Sleep(2 * time.Second)
		continue
	}

}
