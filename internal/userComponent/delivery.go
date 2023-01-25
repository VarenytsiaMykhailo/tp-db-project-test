package userComponent

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

func (d *Delivery) UserCreateHandler(context echo.Context) error {
	var profile *domain.User = &domain.User{}

	err := context.Bind(&profile)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	profile.Nickname = context.Param("nickname")

	_, err = d.db.Exec(`
INSERT INTO profile (nickname, about, email, fullname) 
VALUES ($1, $2, $3, $4);
`, profile.Nickname, profile.About, profile.Email, profile.Fullname)
	if err == nil {

		return context.JSON(201, profile)
	}

	rows, err := d.db.Query(`
SELECT profile.nickname, profile.about, profile.email, profile.fullname 
FROM profile 
WHERE profile.nickname = $1 OR profile.email = $2;
`, profile.Nickname, profile.Email)

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	defer func() {
		rows.Close()
	}()

	var existingProfiles []domain.User
	for rows.Next() {
		var existingProfile *domain.User = &domain.User{}
		err := rows.Scan(&existingProfile.Nickname, &existingProfile.About, &existingProfile.Email,
			&existingProfile.Fullname)
		if err != nil {
			panic(err)
		}
		existingProfiles = append(existingProfiles, *existingProfile)
	}

	return context.JSON(409, existingProfiles)
}

func (d *Delivery) UserGetOnHandler(context echo.Context) error {
	var profile *domain.User = &domain.User{}

	row := d.db.QueryRow(`
SELECT profile.nickname, profile.about, profile.email, profile.fullname 
FROM profile 
WHERE profile.nickname = $1;
`, context.Param("nickname"))
	err := row.Scan(&profile.Nickname, &profile.About, &profile.Email, &profile.Fullname)
	if err == sql.ErrNoRows {

		return context.JSON(404, domain.Error{Message: "" + profile.Nickname})
	}

	return context.JSON(200, profile)
}

func (d *Delivery) UserUpdateHandler(context echo.Context) error {
	var profile domain.User

	profile.Nickname = context.Param("nickname")

	row := d.db.QueryRow(`
SELECT profile.nickname, profile.about, profile.email, profile.fullname 
FROM profile 
WHERE profile.nickname = $1;
`, profile.Nickname)
	err := row.Scan(&profile.Nickname, &profile.About, &profile.Email, &profile.Fullname)
	if err == sql.ErrNoRows {
		return context.JSON(404, domain.Error{Message: ""})
	}

	updatedProfile := profile
	err = context.Bind(&updatedProfile)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	err = d.db.QueryRow(`
SELECT profile.nickname 
FROM profile 
WHERE profile.email = $1 AND profile.nickname != $2;`,
		updatedProfile.Email, profile.Nickname).Scan(&updatedProfile.Nickname)
	if err != sql.ErrNoRows {

		return context.JSON(409, domain.Error{Message: ""})
	}

	_, err = d.db.Exec(`
UPDATE profile SET about = $2, email = $3, fullname = $4 WHERE nickname = $1;
`, updatedProfile.Nickname, updatedProfile.About, updatedProfile.Email, updatedProfile.Fullname)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	return context.JSON(200, updatedProfile)
}
