package serviceComponent

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"tp-db-project/internal/domain"
)

type Delivery struct {
	db *sql.DB
}

func NewDelivery(db *sql.DB) *Delivery {
	return &Delivery{db: db}
}

func (d *Delivery) ServiceClearHandler(context echo.Context) error {
	_, err := d.db.Exec(`
TRUNCATE TABLE profile RESTART IDENTITY CASCADE;
`)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	return context.JSON(200, "Очистка базы успешно завершена")
}

func (d *Delivery) ServiceStatusHandler(context echo.Context) error {
	var status *domain.Status = &domain.Status{}
	row := d.db.QueryRow(`
SELECT (SELECT COUNT(*) FROM forum), (SELECT COUNT(*) FROM post), (SELECT COUNT(*) FROM thread), (SELECT COUNT(*) FROM profile);
`)
	err := row.Scan(&status.Forum, &status.Post, &status.Thread, &status.User)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	return context.JSON(200, status)
}
