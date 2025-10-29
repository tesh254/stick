package studio

import (
    "github.com/gofiber/fiber/v2"
    "github.com/tesh254/stick/internal/functions"
)

func registerRoutes(app *fiber.App, cfg Config, rmFactory RepoManagerFactory, frFactory FuncRegistryFactory) {
    api := app.Group("/api")

    // Health
    api.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status": "ok",
            "env":    cfg.Env,
            "version": VersionString(),
        })
    })

    // Conversations
    api.Get("/conversations", conversationsListHandler(rmFactory))
    api.Get("/conversations/:id", conversationGetHandler(rmFactory))
    api.Get("/conversations/:id/messages", conversationMessagesHandler(rmFactory))

    // Usage
    api.Get("/usage/:conversationId", usageByConversationHandler(rmFactory))

    // Functions
    registry := frFactory()
    api.Get("/functions", functionsListHandler(registry))
    api.Post("/functions/:name/execute", functionExecuteHandler(registry))
}

// VersionString returns a simple version identifier (can be swapped later)
func VersionString() string { return "studio-v0" }

// DefaultFuncRegistryFactory registers baseline standard functions
func DefaultFuncRegistryFactory() FuncRegistryFactory {
    return func() *functions.Registry {
        r := functions.NewRegistry()
        r.Register("add", functions.Add, 0, 2)
        r.Register("echo", functions.Echo, 0, -1)
        r.Register("get_llm_text", functions.GetLLMText, 1, 1)
        r.Register("get_page_content", functions.GetPageHTMLContentToMarkdown, 1, 1)
        return r
    }
}