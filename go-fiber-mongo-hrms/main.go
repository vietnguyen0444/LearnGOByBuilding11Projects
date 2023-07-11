package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoInstance struct {
	Client *mongo.Client
	Db     *mongo.Database
}

var mg MongoInstance

const dbName = "fiber-hrms"
const mongoURI = "mongodb://localhost:27017/" + dbName

type Empoyee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name"`
	Salary float64 `json:"salary"`
	Age    float64 `json:"age"`
}

func Connect() error {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	db := client.Database(dbName)

	if err != nil {
		return err
	}

	mg = MongoInstance{
		Client: client,
		Db:     db,
	}

	return nil
}

func main() {

	if err := Connect(); err != nil {
		log.Fatal(err)
	}
	app := fiber.New()

	// GET Method
	app.Get("/empoyees", func(c *fiber.Ctx) error {

		query := bson.D{{}}

		cursor, err := mg.Db.Collection("employees").Find(c.Context(), query)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		var employees []Empoyee = make([]Empoyee, 0)

		if err := cursor.All(c.Context(), &employees); err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.JSON(employees)
	})

	// POST Method
	app.Post("/empoyees", func(c *fiber.Ctx) error {
		collection := mg.Db.Collection("employees")

		employee := new(Empoyee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		employee.ID = ""

		insertionResult, err := collection.InsertOne(c.Context(), employee)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		filter := bson.D{{Key: "_id", Value: insertionResult.InsertedID}}
		createdRecord := collection.FindOne(c.Context(), filter)

		createdEmployee := &Empoyee{}
		createdRecord.Decode(createdEmployee)

		return c.Status(201).JSON(createdEmployee)
	})

	// PUT Method
	app.Put("/empoyees/:id", func(c *fiber.Ctx) error {
		idParam := c.Params("id")

		employeeID, err := primitive.ObjectIDFromHex(idParam)

		if err != nil {
			return c.SendStatus(400)
		}

		employee := new(Empoyee)

		if err := c.BodyParser(employee); err != nil {
			return c.Status(400).SendString(err.Error())
		}

		// Find existing employee by ID
		query := bson.D{{Key: "_id", Value: employeeID}}
		update := bson.D{
			{
				Key: "$set",
				Value: bson.D{
					{Key: "name", Value: employee.Name},
					{Key: "salary", Value: employee.Salary},
					{Key: "age", Value: employee.Age},
				},
			},
		}

		mg.Db.Collection("employees").FindOneAndUpdate(c.Context(), query, update).Err()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.SendStatus(400)
			}
			return c.SendStatus(500)
		}

		employee.ID = idParam

		return c.Status(200).JSON(employee)
	})

	// DELETE Method
	app.Delete("/empoyees/:id", func(c *fiber.Ctx) error {

		employeeID, err := primitive.ObjectIDFromHex(c.Params("id"))

		if err != nil {
			return c.SendStatus(400)
		}

		query := bson.D{{Key: "_id", Value: employeeID}}
		result, err := mg.Db.Collection("employees").DeleteOne(c.Context(), &query)

		if err != nil {
			return c.SendStatus(500)
		}

		if result.DeletedCount < 1 {
			return c.SendStatus(400)
		}

		return c.Status(200).JSON("record deleted")
	})

	log.Fatal(app.Listen(":3004"))
}
