package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/log"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/robfig/cron/v3"
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

type UserClerk struct {
	Data            Data            `json:"data"`
	EventAttributes EventAttributes `json:"event_attributes"`
	InstanceID      string          `json:"instance_id"`
	Object          string          `json:"object"`
	Timestamp       int64           `json:"timestamp"`
	Type            string          `json:"type"`
}
type LinkedTo struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}
type Verification struct {
	Attempts any    `json:"attempts"`
	ExpireAt any    `json:"expire_at"`
	Object   string `json:"object"`
	Status   string `json:"status"`
	Strategy string `json:"strategy"`
}
type EmailAddresses struct {
	CreatedAt            int64        `json:"created_at"`
	EmailAddress         string       `json:"email_address"`
	ID                   string       `json:"id"`
	LinkedTo             []LinkedTo   `json:"linked_to"`
	MatchesSsoConnection bool         `json:"matches_sso_connection"`
	Object               string       `json:"object"`
	Reserved             bool         `json:"reserved"`
	UpdatedAt            int64        `json:"updated_at"`
	Verification         Verification `json:"verification"`
}
type PublicMetadata struct {
}
type ExternalAccounts struct {
	ApprovedScopes       string         `json:"approved_scopes"`
	AvatarURL            string         `json:"avatar_url"`
	CreatedAt            int64          `json:"created_at"`
	EmailAddress         string         `json:"email_address"`
	EmailAddressVerified bool           `json:"email_address_verified"`
	ExternalAccountID    string         `json:"external_account_id"`
	FamilyName           string         `json:"family_name"`
	FirstName            string         `json:"first_name"`
	GivenName            string         `json:"given_name"`
	GoogleID             string         `json:"google_id"`
	ID                   string         `json:"id"`
	IdentificationID     string         `json:"identification_id"`
	Label                any            `json:"label"`
	LastName             string         `json:"last_name"`
	Object               string         `json:"object"`
	Picture              string         `json:"picture"`
	Provider             string         `json:"provider"`
	ProviderUserID       string         `json:"provider_user_id"`
	PublicMetadata       PublicMetadata `json:"public_metadata"`
	UpdatedAt            int64          `json:"updated_at"`
	Username             any            `json:"username"`
	Verification         Verification   `json:"verification"`
}
type PrivateMetadata struct {
}
type UnsafeMetadata struct {
}
type Data struct {
	BackupCodeEnabled             bool               `json:"backup_code_enabled"`
	Banned                        bool               `json:"banned"`
	BypassClientTrust             bool               `json:"bypass_client_trust"`
	CreateOrganizationEnabled     bool               `json:"create_organization_enabled"`
	CreatedAt                     int64              `json:"created_at"`
	DeleteSelfEnabled             bool               `json:"delete_self_enabled"`
	EmailAddresses                []EmailAddresses   `json:"email_addresses"`
	EnterpriseAccounts            []any              `json:"enterprise_accounts"`
	ExternalAccounts              []ExternalAccounts `json:"external_accounts"`
	ExternalID                    any                `json:"external_id"`
	FirstName                     string             `json:"first_name"`
	HasImage                      bool               `json:"has_image"`
	ID                            string             `json:"id"`
	ImageURL                      string             `json:"image_url"`
	LastActiveAt                  int64              `json:"last_active_at"`
	LastName                      string             `json:"last_name"`
	LastSignInAt                  any                `json:"last_sign_in_at"`
	LegalAcceptedAt               any                `json:"legal_accepted_at"`
	Locale                        any                `json:"locale"`
	Locked                        bool               `json:"locked"`
	LockoutExpiresInSeconds       any                `json:"lockout_expires_in_seconds"`
	MfaDisabledAt                 any                `json:"mfa_disabled_at"`
	MfaEnabledAt                  any                `json:"mfa_enabled_at"`
	Object                        string             `json:"object"`
	Passkeys                      []any              `json:"passkeys"`
	PasswordEnabled               bool               `json:"password_enabled"`
	PasswordLastUpdatedAt         any                `json:"password_last_updated_at"`
	PhoneNumbers                  []any              `json:"phone_numbers"`
	PrimaryEmailAddressID         string             `json:"primary_email_address_id"`
	PrimaryPhoneNumberID          any                `json:"primary_phone_number_id"`
	PrimaryWeb3WalletID           any                `json:"primary_web3_wallet_id"`
	PrivateMetadata               PrivateMetadata    `json:"private_metadata"`
	ProfileImageURL               string             `json:"profile_image_url"`
	PublicMetadata                PublicMetadata     `json:"public_metadata"`
	RequiresPasswordReset         bool               `json:"requires_password_reset"`
	SamlAccounts                  []any              `json:"saml_accounts"`
	TotpEnabled                   bool               `json:"totp_enabled"`
	TwoFactorEnabled              bool               `json:"two_factor_enabled"`
	UnsafeMetadata                UnsafeMetadata     `json:"unsafe_metadata"`
	UpdatedAt                     int64              `json:"updated_at"`
	Username                      any                `json:"username"`
	VerificationAttemptsRemaining int                `json:"verification_attempts_remaining"`
	Web3Wallets                   []any              `json:"web3_wallets"`
}
type HTTPRequest struct {
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}
type EventAttributes struct {
	HTTPRequest HTTPRequest `json:"http_request"`
}

type SubItem struct {
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	DueDate string  `json:"dueDate"`
	Cycle   string  `json:"cycle"`
}

type StudyItem struct {
	Name    string `json:"name"`
	DueDate string `json:"dueDate"`
}

// 1. Update the parse function to respect the local timezone
func parseDueDate(dateStr string, loc *time.Location) (time.Time, error) {
	layout := "Jan 02, 2006"
	return time.ParseInLocation(layout, dateStr, loc)
}

func checkAssignments(db *pgxpool.Pool) {
	log.Info("Running assignments check...")
	ctx := context.Background()

	rows, err := db.Query(ctx, "SELECT user_id, content FROM categories WHERE name='Academics'")
	if err != nil {
		log.Error("Failed to query academics: ", err)
		return
	}
	defer rows.Close()

	now := time.Now()
	// 2. Normalize "today" to exactly 00:00:00 (Midnight) to prevent hour-truncation bugs
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for rows.Next() {
		var userId string
		var content json.RawMessage
		if err := rows.Scan(&userId, &content); err != nil {
			continue
		}

		var wrapper map[string][]StudyItem
		if err := json.Unmarshal(content, &wrapper); err != nil {
			continue
		}

		for _, task := range wrapper["items"] {
			// 3. Pass the location so both dates are in the same timezone
			dueDate, err := parseDueDate(task.DueDate, now.Location())
			if err != nil {
				continue
			}

			// 4. Calculate calendar days remaining using our normalized 'today'
			daysUntilDue := int(dueDate.Sub(today).Hours() / 24)

			// Alert if due in 2 days, 1 day, or TODAY
			if daysUntilDue >= 0 && daysUntilDue <= 2 {
				urgency := "URGENT"
				if daysUntilDue == 0 {
					urgency = "DUE TODAY"
				}

				cleanName := strings.ReplaceAll(task.Name, "🔴 ", "")
				cleanName = strings.ReplaceAll(cleanName, "🟡 ", "")
				cleanName = strings.ReplaceAll(cleanName, "🟢 ", "")

				alertMsg := fmt.Sprintf("📚 %s: %s is due in %d days!", urgency, cleanName, daysUntilDue)
				if daysUntilDue == 0 {
					alertMsg = fmt.Sprintf("📚 %s: %s!", urgency, cleanName)
				}

				log.Info("Sending study alert to user: ", userId, " -> ", alertMsg)

				http.Post("https://ntfy.sh/hackaton", "text/plain", strings.NewReader(alertMsg))
			}
		}
	}
}

func checkSubscriptions(db *pgxpool.Pool) {
	log.Info("Running daily subscription check...")
	ctx := context.Background()

	// 1. Fetch all subscription categories from the database
	rows, err := db.Query(ctx, "SELECT user_id, content FROM categories WHERE name='Subscriptions'")
	if err != nil {
		log.Error("Failed to query subscriptions: ", err)
		return
	}
	defer rows.Close()

	now := time.Now()

	// 2. Iterate through every user's subscriptions
	for rows.Next() {
		var userId string
		var content json.RawMessage

		if err := rows.Scan(&userId, &content); err != nil {
			continue
		}

		// Extract the items array
		var wrapper map[string][]SubItem
		if err := json.Unmarshal(content, &wrapper); err != nil {
			continue
		}

		// 3. Check each subscription date
		for _, sub := range wrapper["items"] {
			if sub.DueDate == "TBD" {
				continue
			}

			dueDate, err := parseDueDate(sub.DueDate, now.Location())
			if err != nil {
				log.Errorf("Invalid date format for %s: %s", sub.Name, sub.DueDate)
				continue
			}

			// Calculate the difference in days
			daysUntilDue := int(dueDate.Sub(now).Hours() / 24)

			// 4. Trigger Notifications!
			// If it's due in exactly 3 days, or if it is due TODAY (0 days)
			if daysUntilDue == 3 || daysUntilDue == 0 {
				alertMsg := fmt.Sprintf("Reminder: %s (%s) renews in %d days for $%.2f!", sub.Name, sub.Cycle, daysUntilDue, sub.Price)
				if daysUntilDue == 0 {
					alertMsg = fmt.Sprintf("ALERT: %s renews TODAY for $%.2f!", sub.Name, sub.Price)
				}

				log.Info("Sending alert to user: ", userId, " -> ", alertMsg)

				http.Post("https://ntfy.sh/hackaton", "text/plain", strings.NewReader(alertMsg))
			}
		}
	}
}

func main() {
	var err error
	db, err := connecDatabase()
	if err != nil {
		panic(err)
	}

	c := cron.New()

	//// every 5 seconds
	_, err = c.AddFunc("@every 5m", func() { checkSubscriptions(db) })
	if err != nil {
		return
	}
	_, err = c.AddFunc("@every 5m", func() { checkAssignments(db) })
	if err != nil {
		return
	}

	c.Start()
	defer c.Stop() // Ensure cron stops gracefully when the app closes
	log.Info("Subscription scheduler started.")

	app := fiber.New()

	app.Use(cors.New())

	users := app.Group("/users")
	categories := app.Group("/categories")
	recipes := app.Group("/recipes")
	scrapers := app.Group("/scrapers")

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

	app.Post("/clerk", func(c fiber.Ctx) error {

		var event UserClerk
		if err := c.Bind().Body(&event); err != nil {
			log.Error("Failed to parse Clerk event: ", err)
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		log.Infof("Received Clerk event: %s for user %s", event.Type, event.Data.ID)

		// save into users table if user doesn't exist
		var existingUserId string
		err := db.QueryRow(c.Context(), "SELECT id FROM users WHERE id=$1", event.Data.ID).Scan(&existingUserId)
		if err != nil {
			if err.Error() == "no rows in result set" {
				// user doesn't exist, insert into database
				_, err := db.Exec(c.Context(), "INSERT INTO users (id, name, email) VALUES ($1, $2, $3)", event.Data.ID, fmt.Sprintf("%s %s", event.Data.FirstName, event.Data.LastName), event.Data.EmailAddresses[0].EmailAddress)
				if err != nil {
					log.Error("Failed to insert user into database: ", err)
					return c.Status(500).JSON(fiber.Map{"error": "Failed to save user data " + err.Error()})
				}
				log.Infof("Inserted new user into database: %s", event.Data.ID)
			} else {
				log.Error("Database query error: ", err)
				return c.Status(500).JSON(fiber.Map{"error": "Database error"})
			}
		}

		return c.SendStatus(200)
	})

	// --- UPDATED CANVAS SCRAPER ---
	scrapers.Post("/canvas", func(c fiber.Ctx) error {
		userId := c.Query("user_id")
		if userId == "" {
			return c.Status(400).JSON(fiber.Map{"error": "user_id query parameter is required"})
		}

		// 1. Check if the user already has scraped data
		var existingContent []byte
		err := db.QueryRow(c.Context(), "SELECT content FROM categories WHERE user_id=$1 AND name='Academics'", userId).Scan(&existingContent)

		if err == nil {
			// It already exists! Just return the existing data and skip the scraping delay.
			var wrapper map[string][]StudyItem
			json.Unmarshal(existingContent, &wrapper)
			return c.JSON(fiber.Map{"items": wrapper["items"]})
		}

		// 2. If it doesn't exist, simulate the scraping delay (First time only)
		time.Sleep(2 * time.Second)

		courses := []string{"MATH 201", "HIST 105", "CS 350", "PHYS 101", "ENG 202"}
		tasks := []string{"Final Exam", "Midterm Essay", "Lab Report", "Problem Set 4", "Reading Quiz"}

		numItems := rand.Intn(3) + 2
		var results []StudyItem

		for i := 0; i < numItems; i++ {
			course := courses[rand.Intn(len(courses))]
			task := tasks[rand.Intn(len(tasks))]
			days := rand.Intn(14) + 1

			targetDate := time.Now().AddDate(0, 0, days).Format("Jan 02, 2006")

			color := "🟢"
			if days <= 2 {
				color = "🔴"
			} else if days <= 5 {
				color = "🟡"
			}

			results = append(results, StudyItem{
				Name:    fmt.Sprintf("%s %s: %s", color, course, task),
				DueDate: targetDate,
			})
		}

		// 3. Save it to the database so it's only generated this first time
		contentBytes, _ := json.Marshal(fiber.Map{"items": results})
		_, err = db.Exec(c.Context(), "INSERT INTO categories (user_id, name, content) VALUES ($1, $2, $3)", userId, "Academics", contentBytes)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to save scraped data"})
		}

		return c.JSON(fiber.Map{"items": results})
	})

	log.Fatal(app.Listen(":8080"))
}
