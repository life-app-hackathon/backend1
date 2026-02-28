package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
)

type User struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Category struct {
	Id      string          `json:"id"`
	UserId  string          `json:"user_id"`
	Name    string          `json:"name"`
	Content json.RawMessage `json:"content"`
}

type RecipeRequest struct {
	Ingredients []string `json:"ingredients"`
}

func main() {
	var err error
	db, err := connecDatabase()
	if err != nil {
		panic(err)
	}

	app := fiber.New()

	users := app.Group("/users")
	categories := app.Group("/categories")
	recipes := app.Group("/recipes")

	// --- USERS ENDPOINTS ---
	users.Post("/", func(c fiber.Ctx) error {
		u := new(User)
		if err := c.Bind().Body(u); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if u.Id == "" {
			query := "INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id"
			err = db.QueryRow(c.Context(), query, u.Name, u.Email).Scan(&u.Id)
		} else {
			query := "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)"
			_, err = db.Exec(c.Context(), query, u.Id, u.Name, u.Email)
		}

		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(u)
	})

	// --- CATEGORIES ENDPOINTS ---
	categories.Get("/:user_id", func(c fiber.Ctx) error {
		userId := c.Params("user_id")
		rows, err := db.Query(c.Context(), "SELECT id, name, content FROM categories WHERE user_id=$1", userId)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		defer rows.Close()

		var cats []Category
		for rows.Next() {
			var cat Category
			cat.UserId = userId
			rows.Scan(&cat.Id, &cat.Name, &cat.Content)
			cats = append(cats, cat)
		}

		if cats == nil {
			cats = []Category{}
		}
		return c.JSON(cats)
	})

	categories.Post("/", func(c fiber.Ctx) error {
		cat := new(Category)
		if err := c.Bind().Body(cat); err != nil {
			return err
		}

		if cat.Id == "" {
			query := "INSERT INTO categories (user_id, name, content) VALUES ($1, $2, $3) RETURNING id"
			err := db.QueryRow(c.Context(), query, cat.UserId, cat.Name, cat.Content).Scan(&cat.Id)
			if err != nil {
				return c.Status(500).SendString(err.Error())
			}
		} else {
			query := "INSERT INTO categories (id, user_id, name, content) VALUES ($1, $2, $3, $4)"
			_, err := db.Exec(c.Context(), query, cat.Id, cat.UserId, cat.Name, cat.Content)
			if err != nil {
				return c.Status(500).SendString(err.Error())
			}
		}
		return c.Status(201).JSON(cat)
	})

	categories.Put("/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		cat := new(Category)
		if err := c.Bind().Body(cat); err != nil {
			return err
		}

		query := "UPDATE categories SET content=$1 WHERE id=$2 AND user_id=$3"
		_, err := db.Exec(c.Context(), query, cat.Content, id, cat.UserId)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendStatus(200)
	})

	categories.Delete("/:id", func(c fiber.Ctx) error {
		id := c.Params("id")
		userId := c.Query("user_id")
		_, err := db.Exec(c.Context(), "DELETE FROM categories WHERE id=$1 AND user_id=$2", id, userId)
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.SendStatus(204)
	})

	// --- RECIPES ENDPOINTS ---
	recipes.Post("/generate", func(c fiber.Ctx) error {
		log.Info("Received recipe generation request")
		req := new(RecipeRequest)
		if err := c.Bind().Body(req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		if len(req.Ingredients) == 0 {
			return c.JSON(fiber.Map{
				"recipe": "❌ You haven't selected any ingredients.\nGo back and try again!",
			})
		}

		ings := strings.Join(req.Ingredients, ", ")
		mockRecipe := fmt.Sprintf(
			"🍲 Magic API Generated Recipe\n\n"+
				"Selected ingredients:\n- %s\n\n"+
				"Instructions:\n"+
				"1. Chop everything into small pieces.\n"+
				"2. Heat a pan with a little olive oil.\n"+
				"3. Sauté the ingredients over medium heat for 10 minutes.\n"+
				"4. Season with salt, pepper, and your favorite spices.\n"+
				"5. Serve hot and enjoy your creation!", ings)

		return c.JSON(fiber.Map{"recipe": mockRecipe})
	})

	log.Fatal(app.Listen(":8080"))
}
