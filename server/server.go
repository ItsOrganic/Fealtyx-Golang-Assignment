package server

import (
	"github.com/ItsOrganic/FealtyX-GoLang_Assignment/handler"
	"github.com/gin-gonic/gin"
)

func Init() {
	r := gin.Default()
	handler := &handler.Handler{}

	r.POST("/student", handler.CreateStudent)
	r.GET("/student/:id", handler.GetStudent)
	r.GET("/students", handler.GetStudents)
	r.PUT("/student/:id", handler.UpdateStudent)
	r.DELETE("/student/:id", handler.DeleteStudent)
	r.GET("/student/:id/summary", handler.Summary)

	r.Run()
}
