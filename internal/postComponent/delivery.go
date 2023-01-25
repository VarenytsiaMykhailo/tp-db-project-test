package postComponent

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	"strconv"
	"strings"
	"tp-db-project/internal/domain"
)

type Delivery struct {
	db *sql.DB
}

func NewDelivery(db *sql.DB) *Delivery {
	return &Delivery{db: db}
}

func (d *Delivery) PostGetOneHandler(context echo.Context) error {
	var postFull *domain.PostFull = &domain.PostFull{}

	id := context.Param("id")
	if id == "" {
		fmt.Println("id is empty")
		panic("id is empty")
	}

	postFull.Post.Id, _ = strconv.ParseUint(id, 10, 64)
	var user, forum, thread bool
	var bitFlag uint8

	for _, related := range strings.Split(context.QueryParam("related"), ",") {
		switch related {
		case "user":
			user = true
			bitFlag |= 1
			break
		case "forum":
			forum = true
			bitFlag |= 2
		case "thread":
			thread = true
			bitFlag |= 4
			break
		}
	}

	var parentPostId sql.NullInt64

	row := d.db.QueryRow(`
SELECT post.profile_nickname, post.created, post.is_edited, post.message, post.post_parent_id, post.thread_id, post.forum_slug 
FROM post 
WHERE post.id = $1;
`, postFull.Post.Id)
	err := row.Scan(&postFull.Post.ProfileNickname, &postFull.Post.Created, &postFull.Post.IsEdited,
		&postFull.Post.Message, &parentPostId, &postFull.Post.ThreadId, &postFull.Post.ForumSlug)
	if err != nil {
		return context.JSON(404, domain.Error{
			Message: "",
		})
	}
	if parentPostId.Valid {
		postFull.Post.ParentPost = uint64(parentPostId.Int64)
	}

	if user {
		postFull.Profile = &domain.User{}

		row := d.db.QueryRow(`
SELECT profile.nickname, profile.about, profile.email, profile.fullname 
FROM profile 
WHERE profile.nickname = $1;
`, postFull.Post.ProfileNickname)
		err := row.Scan(&postFull.Profile.Nickname, &postFull.Profile.About, &postFull.Profile.Email,
			&postFull.Profile.Fullname)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
	}

	if forum {
		postFull.Forum = &domain.Forum{}

		row := d.db.QueryRow(`
SELECT forum.slug, forum.title, forum.profile_nickname, forum.threads, forum.posts 
FROM forum 
WHERE forum.slug = $1;
`, postFull.Post.ForumSlug)
		err := row.Scan(&postFull.Forum.Slug, &postFull.Forum.Title, &postFull.Forum.ProfileNickname, &postFull.Forum.Threads,
			&postFull.Forum.Posts)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
	}

	if thread {
		postFull.Thread = &domain.Thread{}
		var threadSlug sql.NullString
		row := d.db.QueryRow(`
SELECT thread.id, thread.profile_nickname, thread.created, thread.forum_slug, thread.message, thread.slug, thread.title, thread.votes 
FROM thread 
WHERE thread.id = $1;

`, postFull.Post.ThreadId)
		err := row.Scan(&postFull.Thread.Id, &postFull.Thread.ProfileNickname, &postFull.Thread.Created,
			&postFull.Thread.ForumSlug, &postFull.Thread.Message, &threadSlug, &postFull.Thread.Title,
			&postFull.Thread.Votes)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}

		if threadSlug.Valid {
			postFull.Thread.Slug = threadSlug.String
		}
	}

	return context.JSON(200, postFull)
}

func (d *Delivery) PostUpdateHandler(context echo.Context) error {
	var post domain.Post

	id := context.Param("id")
	post.Id, _ = strconv.ParseUint(id, 10, 64)

	row := d.db.QueryRow(`
SELECT post.id, post.profile_nickname, post.created, post.is_edited, post.message, post.thread_id, post.forum_slug
FROM post
WHERE post.id = $1;
`, post.Id)

	err := row.Scan(&post.Id, &post.ProfileNickname, &post.Created, &post.IsEdited, &post.Message, &post.ThreadId,
		&post.ForumSlug)
	if err != nil {
		return context.JSON(404, domain.Error{Message: ""})
	}

	updatedPost := post
	err = context.Bind(&updatedPost)
	if err != nil {
		panic(err)
	}

	if updatedPost.Message != post.Message {

		_, err := d.db.Exec(`
UPDATE post SET message = $1 WHERE id = $2;
`, updatedPost.Message, updatedPost.Id)
		if err != nil {
			fmt.Println(err.Error())
			panic(err)
		}
		updatedPost.IsEdited = true
	}

	return context.JSON(200, updatedPost)
}
