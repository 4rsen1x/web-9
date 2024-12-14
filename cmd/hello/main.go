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

func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, msg)
}

func (h *Handlers) PostHello(c echo.Context) error {
	input := struct {
		Msg string `json:"msg"`
	}{}

	if err := c.Bind(&input); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	if err := h.dbProvider.InsertHello(input.Msg); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	err := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1").Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
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
	CREATE TABLE IF NOT EXISTS hello (
		id SERIAL PRIMARY KEY,
		message TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем маршруты
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)

	// Запускаем сервер
	err = e.Start(*address)
	if err != nil {
		log.Fatal(err)
	}
}
