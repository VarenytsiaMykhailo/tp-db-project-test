package forumComponent

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

func (d *Delivery) ForumCreateHandler(context echo.Context) error {
	var forum *domain.Forum

	if err := context.Bind(&forum); err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}

	if err := d.db.QueryRow(`
INSERT INTO forum (slug, title, profile_nickname) 
SELECT $1, $2, profile.nickname 
FROM profile 
WHERE profile.nickname = $3 RETURNING forum.profile_nickname;`,
		forum.Slug, forum.Title, forum.ProfileNickname).Scan(&forum.ProfileNickname); err != nil {
		if err == sql.ErrNoRows {
			return context.JSON(404, domain.Error{Message: ""})
		} else {
			row := d.db.QueryRow(`
SELECT forum.slug, forum.title, forum.profile_nickname, forum.posts, forum.threads FROM forum WHERE forum.slug = $1;
`, forum.Slug)
			err = row.Scan(&forum.Slug, &forum.Title, &forum.ProfileNickname, &forum.Posts, &forum.Threads)
			if err != nil {
				fmt.Printf(err.Error())
				panic(err)
			}

			return context.JSON(409, forum)
		}
	}

	return context.JSON(201, forum)
}

func (d *Delivery) ForumGetOneHandler(context echo.Context) error {
	var forum *domain.Forum = &domain.Forum{}

	slug := context.Param("slug")
	if slug == "" {
		fmt.Printf("err, slug is empty")
		panic("err, slug is empty")
	}
	forum.Slug = slug

	row := d.db.QueryRow(`
SELECT forum.slug, forum.title, forum.profile_nickname, forum.posts, forum.threads FROM forum WHERE forum.slug = $1;
`, forum.Slug)
	err := row.Scan(&forum.Slug, &forum.Title, &forum.ProfileNickname, &forum.Posts, &forum.Threads)
	if err == sql.ErrNoRows {
		return context.JSON(404, domain.Error{Message: ""})
	}

	return context.JSON(200, forum)
}

func (d *Delivery) ThreadCreateHandler(context echo.Context) error {
	var thread *domain.Thread
	if err := context.Bind(&thread); err != nil {
		fmt.Printf(err.Error())
		panic(err)
	}

	slug_ := context.Param("slug_")
	if slug_ == "" {
		fmt.Printf("err, slug_ is empty")
		panic("err, slug_ is empty")
	}
	thread.ForumSlug = slug_

	row := d.db.QueryRow(`
INSERT INTO thread (title, profile_nickname, forum_slug,  message, slug, created) 
SELECT $1, profile.nickname, forum.slug,  $3, $4, $5 
FROM profile, forum 
WHERE profile.nickname = $2 AND forum.slug = $6 
RETURNING thread.id, thread.profile_nickname, thread.forum_slug;
`, thread.Title, thread.ProfileNickname, thread.Message, thread.Slug, thread.Created, thread.ForumSlug)
	err := row.Scan(&thread.Id, &thread.ProfileNickname, &thread.ForumSlug)
	if err != nil {
		if err == sql.ErrNoRows {
			return context.JSON(404, domain.Error{Message: ""})
		} else {
			row2 := d.db.QueryRow(
				`SELECT thread.id, thread.title, thread.profile_nickname, thread.forum_slug, thread.message, thread.votes, thread.slug, thread.created    FROM thread WHERE thread.slug = $1;
`, thread.Slug)
			err = row2.Scan(&thread.Id, &thread.Title, &thread.ProfileNickname, &thread.ForumSlug, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
			if err != nil {
				fmt.Printf(err.Error())
				panic(err)
			}

			return context.JSON(409, thread)
		}
	}

	return context.JSON(201, thread)
}

func (d *Delivery) ForumGetUsersHandler(context echo.Context) error {
	limit := context.QueryParam("limit")
	if limit == "" {
		limit = "100"
	}

	var forum *domain.Forum = &domain.Forum{}
	slug := context.Param("slug")
	if slug == "" {
		fmt.Println("slug is empty")
		panic("slug is empty")
	}
	forum.Slug = slug
	row := d.db.QueryRow(`
SELECT 1 FROM forum WHERE forum.slug = $1;
`, forum.Slug)
	err := row.Scan()
	if err == sql.ErrNoRows {
		return context.JSON(404, domain.Error{Message: ""})
	}

	var profiles = make([]domain.User, 0)
	var rows *sql.Rows
	since := context.QueryParam("since")
	desc := context.QueryParam("desc")
	if desc == "true" {

		if since == "" {
			rows, err = d.db.Query(`
SELECT forum_user.profile_nickname, forum_user.profile_about, forum_user.profile_email, forum_user.profile_fullname 
FROM forum_user 
WHERE forum_user.forum_slug = $1 
ORDER BY forum_user.profile_nickname DESC 
LIMIT $2;
`, forum.Slug, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		} else {
			rows, err = d.db.Query(`
SELECT forum_user.profile_nickname, forum_user.profile_about, forum_user.profile_email, forum_user.profile_fullname 
FROM forum_user 
WHERE forum_user.forum_slug = $1 
  AND forum_user.profile_nickname < $2 
ORDER BY forum_user.profile_nickname DESC 
LIMIT $3;
`, forum.Slug, since, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		}

	} else {
		if since == "" {
			rows, err = d.db.Query(`
SELECT forum_user.profile_nickname, forum_user.profile_about, forum_user.profile_email, forum_user.profile_fullname 
FROM forum_user 
WHERE forum_user.forum_slug = $1 
ORDER BY forum_user.profile_nickname 
LIMIT $2;
`, forum.Slug, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		} else {
			rows, err = d.db.Query(`
SELECT forum_user.profile_nickname, forum_user.profile_about, forum_user.profile_email, forum_user.profile_fullname 
FROM forum_user 
WHERE forum_user.forum_slug = $1 AND forum_user.profile_nickname > $2 
ORDER BY forum_user.profile_nickname 
LIMIT $3;
`, forum.Slug, since, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		}

	}

	for rows.Next() {
		var user *domain.User = &domain.User{}
		err := rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		profiles = append(profiles, *user)
	}

	return context.JSON(200, profiles)
}

func (d *Delivery) ForumGetThreadsHandler(context echo.Context) error {
	limit := context.QueryParam("limit")
	if limit == "" {
		limit = "100"
	}

	var forum *domain.Forum = &domain.Forum{}
	forum.Slug = context.Param("slug")
	row := d.db.QueryRow(`
SELECT forum.slug FROM forum WHERE forum.slug = $1;
`, forum.Slug)
	var err error
	err = row.Scan(&forum.Slug)
	if err == sql.ErrNoRows {
		return context.JSON(404, domain.Error{Message: ""})
	}

	var rows *sql.Rows
	since := context.QueryParam("since")
	if context.QueryParam("desc") == "true" {
		if since == "" {
			rows, err = d.db.Query(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.forum_slug = $1 
ORDER BY thread.created
DESC LIMIT $2;
`, forum.Slug, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		} else {
			rows, err = d.db.Query(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.forum_slug = $1 AND thread.created <= $2 
ORDER BY thread.created 
DESC LIMIT $3;
`, forum.Slug, since, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		}
	} else {
		if since == "" {
			rows, err = d.db.Query(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.forum_slug = $1 
ORDER BY thread.created 
LIMIT $2;
`, forum.Slug, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		} else {
			rows, err = d.db.Query(`SELECT thread.id, thread.profile_nickname, thread.created, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.forum_slug = $1 AND thread.created >= $2 
ORDER BY thread.created 
LIMIT $3;
`, forum.Slug, since, limit)
			defer func() {
				rows.Close()
			}()
			if err != nil {
				fmt.Println(err.Error())
				panic(err)
			}
		}
	}

	var threads = make([]domain.Thread, 0)

	for rows.Next() {
		var thread *domain.Thread = &domain.Thread{}
		var threadSlug sql.NullString
		err = rows.Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.Message, &threadSlug,
			&thread.Title, &thread.Votes)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}

		if threadSlug.Valid {
			thread.Slug = threadSlug.String
		}
		thread.ForumSlug = forum.Slug

		threads = append(threads, *thread)
	}

	return context.JSON(200, threads)
}
