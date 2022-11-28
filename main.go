package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
	"todolist.go/db"
	"todolist.go/service"
)

const port = 8000

func main() {
	// initialize DB connection
	dsn := db.DefaultDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"))
	if err := db.Connect(dsn); err != nil {
		log.Fatal(err)
	}

	// initialize Gin engine
	engine := gin.Default()
	engine.LoadHTMLGlob("views/*.html")

	// prepare session
    store := cookie.NewStore([]byte("my-secret"))
    engine.Use(sessions.Sessions("user-session", store))

	// routing
	engine.Static("/assets", "./assets")
	engine.GET("/", service.Home)
	engine.GET("/list", service.LoginCheck, service.TaskList)
	engine.GET("/login", service.LoginForm)
	engine.GET("/logout", service.LoginCheck, service.Logout)
	engine.GET("/task/:id", service.LoginCheckTaskID, service.ShowTask) // ":id" is a parameter

	// タスクの新規登録
    engine.GET("/task/new", service.LoginCheck, service.NewTaskForm)
    engine.POST("/task/new", service.LoginCheck, service.RegisterTask)
    // 既存タスクの編集
    engine.GET("/task/edit/:id", service.LoginCheckTaskID, service.EditTaskForm)
    engine.POST("/task/edit/:id", service.LoginCheckTaskID, service.UpdateTask)
    // 既存タスクの削除
    engine.GET("/task/delete/:id", service.LoginCheckTaskID, service.DeleteTask)


	// ユーザ登録
    engine.GET("/user/new", service.NewUserForm)
    engine.POST("/user/new", service.RegisterUser)

	//アカウント編集
	engine.GET("/user/edit", service.EditUserForm)
	engine.POST("/user/edit", service.UpdateUser)

	//アカウント削除
	engine.GET("/user/delete", service.ResureDelete)
	engine.POST("/user/delete", service.DeleteUser)

	//ログイン
	engine.POST("/login", service.Login)

	// start server
	engine.Run(fmt.Sprintf(":%d", port))
}
