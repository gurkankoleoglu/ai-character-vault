package main

import (
	"stajyer-backend/config"
	"stajyer-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitDatabase()
	r := gin.Default()
	routes.SetupRoutes(r)
	r.Run(":8080")
}
