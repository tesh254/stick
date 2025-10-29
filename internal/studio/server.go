package studio

import (
    "fmt"
    "strings"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/tesh254/stick/internal/db"
)

// Start initializes and runs the Fiber server with configured middleware and routes
func Start(cfg Config) error {
    app := fiber.New(fiber.Config{
        DisableStartupMessage: true,
        // Increase read buffer to accommodate large request headers (e.g., cookies)
        ReadBufferSize: 64 * 1024, // 64KB
    })

    // Custom startup logs
    fmt.Printf("🧪 Stick Studio starting | env=%s | port=%s | origins=%s\n", cfg.Env, cfg.Port, cfg.AllowedOrigins)

    // Recover middleware to protect against panics
    app.Use(recover.New())

    // Custom request logging middleware (replace default Fiber logger)
    app.Use(func(c *fiber.Ctx) error {
        start := time.Now()
        err := c.Next()
        dur := time.Since(start)
        fmt.Printf("➡ %s %s | %d | %s | ip=%s\n", c.Method(), c.OriginalURL(), c.Response().StatusCode(), dur.Round(time.Millisecond), c.IP())
        return err
    })

    // CORS configuration
    app.Use(cors.New(cors.Config{
        AllowOrigins: cfg.AllowedOrigins,
        AllowMethods: "GET,POST,OPTIONS",
        AllowHeaders: "Origin, Content-Type, Accept",
    }))

    // Attach routes
    rmFactory := func() (db.RepositoryManager, error) {
        dbc, err := db.New()
        if err != nil {
            return nil, err
        }
        return db.NewRepositoryManager(dbc), nil
    }

    registerRoutes(app, cfg, rmFactory, DefaultFuncRegistryFactory())

    // Final banner
    fmt.Println("✅ Stick Studio ready | Visit /api/health for status")
    return app.Listen(":" + strings.TrimSpace(cfg.Port))
}