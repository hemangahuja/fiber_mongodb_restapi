package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Body struct {
	Name  string `json:"name"`
	Marks int    `json:"marks"`
}

func main() {
	app := fiber.New()
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		panic(err)
	}
	coll := client.Database("health_cube").Collection("patients")

	app.Get("/", func(c *fiber.Ctx) error {
		fmt.Println("Endpoint hit")
		return c.SendString("Hello, World 👋!")
	})
	app.Get("/:id", func(c *fiber.Ctx) error {
		id, error := primitive.ObjectIDFromHex(c.Params("id"))
		if error != nil {
			return c.Status(400).SendString("wrong id format")
		}
		res := coll.FindOne(context.TODO(), bson.M{
			"_id": id,
		})
		body := new(Body)
		if res.Decode(body) != nil {
			return c.Status(404).SendString("not found")
		}
		bodyJson, _ := json.Marshal(body)
		return c.SendString(string(bodyJson))
	})
	app.Post("/", func(c *fiber.Ctx) error {
		body := new(Body)

		if error := c.BodyParser(&body); error != nil {
			return c.Status(400).SendString(error.Error())
		}
		if res, error := coll.InsertOne(context.TODO(), body); err != nil {
			return c.Status(400).SendString(error.Error())
		} else {
			fmt.Println(res.InsertedID)
			return c.SendString(fmt.Sprintln(res.InsertedID))
		}
	})
	app.Delete("/:id", func(c *fiber.Ctx) error {
		requestId := c.Params("id")
		parsedId, error := primitive.ObjectIDFromHex(requestId)
		if error != nil {
			return c.Status(400).SendString("wrong id format")
		}
		_, error = coll.DeleteOne(context.TODO(), bson.M{
			"_id": parsedId,
		})
		if error != nil {
			return c.Status(404).SendString("not found")
		}
		return c.SendString("ok")
	})
	app.Put("/:id", func(c *fiber.Ctx) error {
		requestId := c.Params("id")
		parsedId, error := primitive.ObjectIDFromHex(requestId)

		if error != nil {
			return c.Status(400).SendString("wrong id format")
		}
		body := new(Body)
		c.BodyParser(&body)

		_, err := coll.UpdateOne(
			context.TODO(),
			bson.M{"_id": parsedId},
			bson.D{
				{Key: "$set", Value: bson.D{{Key: "name", Value: body.Name}}},
				{Key: "$set", Value: bson.D{{Key: "marks", Value: body.Marks}}},
			},
		)
		if err != nil {
			return c.Status(404).SendString("not found")
		}

		return c.SendString("ok")
	})
	app.Listen(":3000")
}
