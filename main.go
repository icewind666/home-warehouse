package main

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type Item struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
	Expires     time.Time `json:"expires"`
}

func main() {
	db, err := sql.Open("postgres", "postgres://user:password@db/items?sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, warehouse is up and running!")
	})

	e.POST("/items", func(c echo.Context) error {
		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}

		stmt, err := db.Prepare("INSERT INTO items (title, description, quantity, expires) VALUES ($1, $2, $3, $4) RETURNING id")
		if err != nil {
			return err
		}

		var id int
		err = stmt.QueryRow(item.Title, item.Description, item.Quantity, item.Expires).Scan(&id)
		if err != nil {
			return err
		}

		item.ID = id

		return c.JSON(http.StatusCreated, item)
	})

	e.GET("/items/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		var title, description string
		var quantity int
		var expires time.Time

		err = db.QueryRow("SELECT title, description, quantity, expires FROM items WHERE id=$1", id).Scan(&title, &description, &quantity, &expires)
		if err != nil {
			return err
		}

		item := &Item{
			ID:          id,
			Title:       title,
			Description: description,
			Quantity:    quantity,
			Expires:     expires,
		}

		return c.JSON(http.StatusOK, item)
	})

	e.PUT("/items/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		item := new(Item)
		if err := c.Bind(item); err != nil {
			return err
		}

		stmt, err := db.Prepare("UPDATE items SET title=$1, description=$2, quantity=$3, expires=$4 WHERE id=$5")
		if err != nil {
			return err
		}

		res, err := stmt.Exec(item.Title, item.Description, item.Quantity, item.Expires, id)
		if err != nil {
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Item not found")
		}

		item.ID = id

		return c.JSON(http.StatusOK, item)
	})

	e.DELETE("/items/:id", func(c echo.Context) error {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			return err
		}

		res, err := db.Exec("DELETE FROM items WHERE id=$1", id)
		if err != nil {
			return err
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Item not found")
		}

		return c.NoContent(http.StatusNoContent)
	})

	e.Logger.Fatal(e.Start(":8080"))
}
