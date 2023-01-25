package threadComponent

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	"io/ioutil"
	"strconv"
	"time"
	"tp-db-project/internal/domain"
)

type Delivery struct {
	db *sql.DB
}

func NewDelivery(db *sql.DB) *Delivery {
	return &Delivery{db: db}
}

func (d *Delivery) PostsCreateHandler(context echo.Context) error {
	var thread domain.Thread
	var threadSlug sql.NullString
	slugOrId := context.Param("slug_or_id")

	if _, err := strconv.Atoi(slugOrId); err == nil {
		row := d.db.QueryRow(`
SELECT thread.id, thread.slug, thread.forum_slug 
FROM thread 
WHERE thread.id = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &threadSlug, &thread.ForumSlug)

		if err != nil {
			return context.JSON(404, domain.Error{Message: "" + slugOrId})
		}
	} else {
		row := d.db.QueryRow(`
SELECT thread.id, thread.slug, thread.forum_slug 
FROM thread 
WHERE thread.slug = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &threadSlug, &thread.ForumSlug)

		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	}

	if threadSlug.Valid {
		thread.Slug = threadSlug.String
	}
	var posts []*domain.Post
	result, err := ioutil.ReadAll(context.Request().Body)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(result, &posts); err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	if len(posts) == 0 {
		return context.JSON(201, posts)
	}
	location, _ := time.LoadLocation("UTC")
	now := time.Now().In(location).Round(time.Microsecond)

	tx, err := d.db.Begin()
	defer func() {
		_ = tx.Rollback()
	}()
	if err != nil {
		panic(err)
	}

	statement, err := tx.Prepare(`
INSERT INTO post (profile_nickname, created, message, post_parent_id, thread_id, forum_slug) 
SELECT profile.nickname, $2, $3, $4, $5, $6 
FROM profile 
WHERE profile.nickname = $1 
RETURNING post.id;
`)
	defer func() {
		statement.Close()
	}()

	for _, post := range posts {
		if post.Created.IsZero() {
			post.Created = now
		}

		err = statement.QueryRow(post.ProfileNickname, post.Created, post.Message, post.ParentPost, thread.Id, thread.ForumSlug).Scan(&post.Id)
		if err != nil {
			if err == sql.ErrNoRows {
				return context.JSON(404, domain.Error{Message: ""})
			} else {
				return context.JSON(409, domain.Error{Message: ""})
			}
		}
		post.ThreadId = thread.Id
		post.ForumSlug = thread.ForumSlug
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	return context.JSON(201, posts)
}

func (d *Delivery) ThreadGetOneHandler(context echo.Context) error {
	var thread domain.Thread
	var threadSlug sql.NullString
	slugOrId := context.Param("slug_or_id")

	if _, err := strconv.Atoi(slugOrId); err == nil {
		row := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.id = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.ForumSlug, &thread.Message,
			&threadSlug, &thread.Title, &thread.Votes)
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}

	} else {
		row := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.slug = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.ForumSlug, &thread.Message,
			&threadSlug, &thread.Title, &thread.Votes)
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	}
	if threadSlug.Valid {
		thread.Slug = threadSlug.String
	}

	return context.JSON(200, thread)
}

func (d *Delivery) ThreadUpdateHandler(context echo.Context) error {
	var thread domain.Thread
	var threadSlug sql.NullString
	slugOrId := context.Param("slug_or_id")

	if _, err := strconv.Atoi(slugOrId); err == nil {
		row := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.id = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.ForumSlug, &thread.Message, &threadSlug, &thread.Title, &thread.Votes)
		if err != nil {
			return context.JSON(404, domain.Error{Message: "" + slugOrId})
		}

	} else {
		row := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.slug = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.ForumSlug, &thread.Message,
			&threadSlug, &thread.Title, &thread.Votes)
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	}

	if threadSlug.Valid {
		thread.Slug = threadSlug.String
	}

	err := context.Bind(&thread)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	_, err = d.db.Exec(`
UPDATE thread SET message = $2, title = $3 WHERE id = $1;
`, thread.Id, thread.Message, thread.Title)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	return context.JSON(200, thread)
}

func (d *Delivery) ThreadGetPostsHandler(context echo.Context) error {
	limit := context.QueryParam("limit")
	if limit == "" {
		limit = "100"
	}
	since := context.QueryParam("since")
	var desc string
	if context.QueryParam("desc") == "true" {
		desc = "DESC"
	}
	var thread domain.Thread
	slugOrId := context.Param("slug_or_id")
	if _, err := strconv.Atoi(slugOrId); err == nil {
		if err := d.db.QueryRow(`
SELECT thread.id, thread.forum_slug 
FROM thread 
WHERE thread.id = $1;
`, slugOrId).Scan(&thread.Id, &thread.ForumSlug); err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	} else {
		if err := d.db.QueryRow(`
SELECT thread.id, thread.forum_slug 
FROM thread 
WHERE thread.slug = $1;
`, slugOrId).Scan(&thread.Id, &thread.ForumSlug); err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	}

	var rows *sql.Rows
	var err error
	switch context.QueryParam("sort") {
	case "tree":
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf("SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id FROM post WHERE post.thread_id = $1 ORDER BY post.path_ %s, post.created, post.id LIMIT $2;", desc),
				thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post WHERE post.thread_id = $1 AND post.path_ > (SELECT post.path_ FROM post WHERE post.id = $2) 
          ORDER BY post.path_, post.created, post.id 
          LIMIT $3;
`, thread.Id, since, limit)
			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post 
WHERE post.thread_id = $1 AND post.path_ < (SELECT post.path_ FROM post WHERE post.id = $2) 
ORDER BY post.path_ DESC, post.created, post.id 
LIMIT $3;`, thread.Id, since, limit)
			}
		}
		break
	case "parent_tree":
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf("SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id FROM post WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 ORDER BY post.id %s LIMIT $2) ORDER BY post.post_root_id %s, post.path_, post.created, post.id;", desc, desc),
				thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post 
WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 AND post.post_root_id > (SELECT post.post_root_id FROM post WHERE post.id = $2) ORDER BY post.id LIMIT $3) 
ORDER BY post.post_root_id, post.path_, post.created, post.id;
`, thread.Id, since, limit)
			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 AND post.post_root_id < (SELECT post.post_root_id FROM post WHERE post.id = $2) ORDER BY post.id DESC LIMIT $3) 
          ORDER BY post.post_root_id DESC, post.path_, post.created, post.id;
`, thread.Id, since, limit)
			}
		}
		break
	default: //flat
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf("SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id FROM post WHERE post.thread_id = $1 ORDER BY post.created %s, post.id %s LIMIT $2;", desc, desc),
				thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post 
WHERE post.thread_id = $1 AND post.id > $2 ORDER BY post.created, post.id LIMIT $3;
`, thread.Id, since, limit)
			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id 
FROM post 
WHERE post.thread_id = $1 AND post.id < $2 ORDER BY post.created DESC, post.id DESC LIMIT $3;
`, thread.Id, since, limit)
			}
		}
		break
	}
	if err != nil {
		panic(err)
	}
	defer func() {
		rows.Close()
	}()

	var posts = make([]domain.Post, 0)
	for rows.Next() {
		var post domain.Post
		var parentPostId sql.NullInt64
		if err := rows.Scan(&post.Id, &post.ProfileNickname, &post.Created, &post.IsEdited, &post.Message,
			&parentPostId); err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		if parentPostId.Valid {
			post.ParentPost = uint64(parentPostId.Int64)
		}
		post.ForumSlug = thread.ForumSlug
		post.ThreadId = thread.Id
		posts = append(posts, post)
	}

	return context.JSON(200, posts)
}

/*
func (d *Delivery) ThreadGetPostsHandler(context echo.Context) error {
	limit := context.QueryParam("limit")
	if limit == "" {
		limit = "100"
	}
	since := context.QueryParam("since")
	var desc string
	if context.QueryParam("desc") == "true" {
		desc = "DESC"
	}

	var thread domain.Thread
	slugOrId := context.Param("slug_or_id")
	if _, err := strconv.Atoi(slugOrId); err == nil {
		row := d.db.QueryRow(`
SELECT thread.id, thread.forum_slug
FROM thread
WHERE thread.id = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ForumSlug)
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}

	} else {

		row := d.db.QueryRow(`
SELECT thread.id, thread.forum_slug
FROM thread
WHERE thread.slug = $1;
`, slugOrId)
		err := row.Scan(&thread.Id, &thread.ForumSlug)
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	}

	var rows *sql.Rows
	var err error

	switch context.QueryParam("sort") {
	case "tree":
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1
ORDER BY post.path_ %s, post.created, post.id
LIMIT $2;
`, desc), thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1 AND post.path_ > (SELECT post.path_ FROM post WHERE post.id = $2)
ORDER BY post.path_, post.created, post.id
LIMIT $3;
`, thread.Id, since, limit)
			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1 AND post.path_ < (SELECT post.path_ FROM post WHERE post.id = $2)
ORDER BY post.path_ DESC, post.created, post.id LIMIT $3;
`, thread.Id, since, limit)
			}
		}
		break
	case "parent_tree":
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 ORDER BY post.id %s LIMIT $2)
ORDER BY post.post_root_id %s, post.path_, post.created, post.id;
`, desc, desc), thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 AND post.post_root_id > (SELECT post.post_root_id FROM post WHERE post.id = $2) ORDER BY post.id LIMIT $3)
ORDER BY post.post_root_id, post.path_, post.created, post.id;
`, thread.Id, since, limit)

			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.post_root_id IN (SELECT post.id FROM post WHERE post.post_parent_id IS NULL AND post.thread_id = $1 AND post.post_root_id < (SELECT post.post_root_id FROM post WHERE post.id = $2) ORDER BY post.id DESC LIMIT $3)
ORDER BY post.post_root_id DESC, post.path_, post.created, post.id;
`, thread.Id, since, limit)
			}
		}
		break
	default:
		if since == "" {
			rows, err = d.db.Query(fmt.Sprintf(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1
ORDER BY post.created %s, post.id %s LIMIT $2;
`, desc, desc), thread.Id, limit)
		} else {
			if desc == "" {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1 AND post.id > $2 ORDER BY post.created, post.id LIMIT $3;
`,
					thread.Id, since, limit)
			} else {
				rows, err = d.db.Query(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id
FROM post
WHERE post.thread_id = $1 AND post.id < $2
ORDER BY post.created DESC, post.id DESC LIMIT $3;
`, thread.Id, since, limit)
			}
		}

		break
	}

	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	defer func() {
		panic(err)
	}()

	var posts = make([]domain.Post, 0)

	for rows.Next() {
		var post domain.Post
		var parentPostId sql.NullInt64
		err := rows.Scan(&post.Id, &post.ProfileNickname, &post.Created, &post.IsEdited, &post.Message,
			&parentPostId)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		if parentPostId.Valid {
			post.ParentPost = uint64(parentPostId.Int64)
		}
		post.ForumSlug = thread.ForumSlug
		post.ThreadId = thread.Id
		posts = append(posts, post)
	}

	return context.JSON(200, posts)
}

*/

func (d *Delivery) ThreadVoteHandler(context echo.Context) error {
	var vote domain.Vote
	if err := context.Bind(&vote); err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	var thread domain.Thread
	slugOrId := context.Param("slug_or_id")
	if _, err := strconv.Atoi(slugOrId); err == nil {
		err := d.db.QueryRow(`
INSERT INTO vote (profile_id, thread_id, voice) 
SELECT profile.id, thread.id, $3 
FROM profile, thread 
WHERE profile.nickname = $1 AND thread.id = $2 ON CONFLICT (profile_id, thread_id) DO UPDATE SET voice = $3 
                                               RETURNING vote.thread_id;
`, vote.ProfileNickname, slugOrId, vote.Voice).Scan(&vote.ThreadId);
		if err != nil {
			return context.JSON(404, domain.Error{Message: ""})
		}
	} else {
		if err = d.db.QueryRow(`
INSERT INTO vote (profile_id, thread_id, voice) 
SELECT profile.id, thread.id, $3 
FROM profile, thread 
WHERE profile.nickname = $1 AND thread.slug = $2 ON CONFLICT (profile_id, thread_id) DO UPDATE SET voice = $3 
                                                 RETURNING vote.thread_id;
`, vote.ProfileNickname, slugOrId, vote.Voice).Scan(&vote.ThreadId); err != nil {

			return context.JSON(404, domain.Error{Message: ""})
		}
	}

	var threadSlug sql.NullString
	if err := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.id = $1;
`, vote.ThreadId).Scan(&thread.Id, &thread.ProfileNickname, &thread.Created, &thread.ForumSlug, &thread.Message, &threadSlug, &thread.Title, &thread.Votes); err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
	if threadSlug.Valid {
		thread.Slug = threadSlug.String
	}

	return context.JSON(200, thread)
}
