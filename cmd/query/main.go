package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "1"
	dbname   = "postgres"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

func (h *Handlers) GetUser(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		name = "Guest"
	}

	err := h.dbProvider.InsertUser(name)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("Hello, %s!", name))
}

func (dp *DatabaseProvider) InsertUser(name string) error {
	_, err := dp.db.Exec("INSERT INTO users (name) VALUES ($1)", name)
	return err
}

func main() {
	// Считываем аргументы командной строки
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}
	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	// Создаем таблицы, если их нет
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем маршруты
	e.GET("/api/user", h.GetUser)

	// Запускаем сервер
	err = e.Start(*address)
	if err != nil {
		log.Fatal(err)
	}
}
