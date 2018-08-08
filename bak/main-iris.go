package main

import (
	"github.com/kataras/iris"
)

type User struct {
	Firstname string `json:"firstname"`
	Age       int    `json:"age"`
}

func main() {
	app := iris.New()

	app.Get("/ping", func(ctx iris.Context) {
		peter := User{
			Firstname: "John",
			Age:       25,
		}
		ctx.JSON(peter)
	})

	app.Run(iris.Addr(":3000"))
}
