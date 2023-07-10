package main

import (
	"fmt"

	"github.com/gofiber/fiber"
	"github.com/jinzhu/gorm"
	"github.com/vietnuyen/go-fiber-crm-basic/database"
	"github.com/vietnuyen/go-fiber-crm-basic/lead"
)

func setupRoutes(app *fiber.App) {
	app.Get("/api/v1/leads", lead.GetLeads)
	app.Get("/api/v1/leads/:id", lead.GetLead)
	app.Post("/api/v1/leads/:id", lead.NewLead)
	app.Delete("/api/v1/leads/:id", lead.DeleteLead)
}

func initDatabase() {
	var err error
	database.DBConn, err = gorm.Open("sqlite3", "leads.db")
	if err != nil {
		panic("Failed to connect database")
	}
	fmt.Println("Connection opened to database")
	database.DBConn.AutoMigrate(&lead.Lead{})
	fmt.Println("Database Migrated")
}

func main() {
	app := fiber.New()
	initDatabase()
	setupRoutes(app)

	app.Listen(3003)
	defer database.DBConn.Close()
}
