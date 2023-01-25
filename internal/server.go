package internal

import (
	"database/sql"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"strconv"
	"tp-db-project/internal/forumComponent"
	"tp-db-project/internal/postComponent"
	"tp-db-project/internal/serviceComponent"
	"tp-db-project/internal/threadComponent"
	"tp-db-project/internal/userComponent"
)

var (
	serverAddress = "0.0.0.0:5000"
)

func Run() {
	db, err := GetPostgreSQLConnections("db_pg", "forums", "admin", "localhost", "5432", "80")
	defer func() {
		_ = db.Close()
	}()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	// На net http было бы быстрее
	e := echo.New()

	forumDelivery := forumComponent.NewDelivery(db)
	postDelivery := postComponent.NewDelivery(db)
	serviceDelivery := serviceComponent.NewDelivery(db)
	threadDelivery := threadComponent.NewDelivery(db)
	userDelivery := userComponent.NewDelivery(db)

	e.POST("/api/forum/create", forumDelivery.ForumCreateHandler)
	e.GET("/api/forum/:slug/details", forumDelivery.ForumGetOneHandler)
	e.POST("/api/forum/:slug_/create", forumDelivery.ThreadCreateHandler)
	e.GET("/api/forum/:slug/users", forumDelivery.ForumGetUsersHandler)
	e.GET("/api/forum/:slug/threads", forumDelivery.ForumGetThreadsHandler)

	e.GET("/api/post/:id/details", postDelivery.PostGetOneHandler)
	e.POST("/api/post/:id/details", postDelivery.PostUpdateHandler)

	e.POST("/api/service/clear", serviceDelivery.ServiceClearHandler)
	e.GET("/api/service/status", serviceDelivery.ServiceStatusHandler)

	e.POST("/api/thread/:slug_or_id/create", threadDelivery.PostsCreateHandler)
	e.GET("/api/thread/:slug_or_id/details", threadDelivery.ThreadGetOneHandler)
	e.POST("/api/thread/:slug_or_id/details", threadDelivery.ThreadUpdateHandler)
	e.GET("/api/thread/:slug_or_id/posts", threadDelivery.ThreadGetPostsHandler)
	e.POST("/api/thread/:slug_or_id/vote", threadDelivery.ThreadVoteHandler)

	e.POST("/api/user/:nickname/create", userDelivery.UserCreateHandler)
	e.GET("/api/user/:nickname/profile", userDelivery.UserGetOnHandler)
	e.POST("/api/user/:nickname/profile", userDelivery.UserUpdateHandler)

	if err := e.Start(serverAddress); err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func GetPostgreSQLConnections(databaseUser string, databaseName string, databasePassword string, databaseHost string, databasePort string, databaseMaxOpenConnections string) (*sql.DB, error) {
	dsn := "user=" + databaseUser + " dbname=" + databaseName + " password=" + databasePassword + " host=" + databaseHost + " port=" + databasePort + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	// Test connection
	if err = db.Ping(); err != nil {
		return nil, err
	}

	databaseMaxOpenConnectionsINT, err := strconv.Atoi(databaseMaxOpenConnections)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(databaseMaxOpenConnectionsINT)
	db.SetMaxIdleConns(databaseMaxOpenConnectionsINT - 2)

	return db, nil
}
