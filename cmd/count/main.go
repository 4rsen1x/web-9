package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

	// "net/http"
	"strconv"

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

func (h *Handlers) GetCount(c echo.Context) error {
	count, err := h.dbProvider.SelectCount()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, strconv.Itoa(count))
}

func (h *Handlers) PostCount(c echo.Context) error {
	countStr := c.FormValue("count")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		return c.String(http.StatusBadRequest, "это не число")
	}

	err = h.dbProvider.UpdateCount(count)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (dp *DatabaseProvider) SelectCount() (int, error) {
	var count int

	err := dp.db.QueryRow("SELECT value FROM counter WHERE id = 1").Scan(&count)
	if err == sql.ErrNoRows {
		_, err = dp.db.Exec("INSERT INTO counter (id, value) VALUES (1, 0)")
		if err != nil {
			return 0, err
		}
		count = 0
	} else if err != nil {
		return 0, err
	}

	return count, nil
}

func (dp *DatabaseProvider) UpdateCount(value int) error {
	_, err := dp.db.Exec("UPDATE counter SET value = value + $1 WHERE id = 1", value)
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
	CREATE TABLE IF NOT EXISTS counter (
		id SERIAL PRIMARY KEY,
		value INTEGER NOT NULL
	);`)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем маршруты
	e.GET("/count", h.GetCount)
	e.POST("/count", h.PostCount)

	// Запускаем сервер
	err = e.Start(*address)
	if err != nil {
		log.Fatal(err)
	}
}
