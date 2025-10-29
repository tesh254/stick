package studio

import (
    "fmt"

    "github.com/dombox/uuidv7"
    "github.com/gofiber/fiber/v2"
    "github.com/tesh254/stick/internal/functions"
)

// conversationsListHandler returns paginated conversations
func conversationsListHandler(rmFactory RepoManagerFactory) fiber.Handler {
    return func(c *fiber.Ctx) error {
        rm, err := rmFactory()
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "db_error", err.Error())
        }
        limit := c.QueryInt("limit", 50)
        offset := c.QueryInt("offset", 0)
        items, err := rm.Conversations().GetAll(c.Context(), limit, offset)
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "query_failed", err.Error())
        }
        return c.JSON(fiber.Map{"data": items, "limit": limit, "offset": offset})
    }
}

// conversationGetHandler returns a conversation with or without messages
func conversationGetHandler(rmFactory RepoManagerFactory) fiber.Handler {
    return func(c *fiber.Ctx) error {
        rm, err := rmFactory()
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "db_error", err.Error())
        }
        idStr := c.Params("id")
        id, err := uuidv7.Parse(idStr)
        if err != nil {
            return apiError(c, fiber.StatusBadRequest, "invalid_id", "invalid uuid")
        }

        withMessages := c.Query("with_messages") == "true"
        if withMessages {
            conv, msgs, err := rm.Conversations().GetWithMessages(c.Context(), id)
            if err != nil {
                return apiError(c, fiber.StatusNotFound, "not_found", err.Error())
            }
            return c.JSON(fiber.Map{"conversation": conv, "messages": msgs})
        }

        conv, err := rm.Conversations().GetByID(c.Context(), id)
        if err != nil {
            return apiError(c, fiber.StatusNotFound, "not_found", err.Error())
        }
        return c.JSON(conv)
    }
}

// conversationMessagesHandler returns messages for a conversation
func conversationMessagesHandler(rmFactory RepoManagerFactory) fiber.Handler {
    return func(c *fiber.Ctx) error {
        rm, err := rmFactory()
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "db_error", err.Error())
        }
        idStr := c.Params("id")
        id, err := uuidv7.Parse(idStr)
        if err != nil {
            return apiError(c, fiber.StatusBadRequest, "invalid_id", "invalid uuid")
        }
        msgs, err := rm.Messages().GetByConversationID(c.Context(), id)
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "query_failed", err.Error())
        }
        return c.JSON(fiber.Map{"data": msgs})
    }
}

// usageByConversationHandler returns usage records by conversation id
func usageByConversationHandler(rmFactory RepoManagerFactory) fiber.Handler {
    return func(c *fiber.Ctx) error {
        rm, err := rmFactory()
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "db_error", err.Error())
        }
        idStr := c.Params("conversationId")
        id, err := uuidv7.Parse(idStr)
        if err != nil {
            return apiError(c, fiber.StatusBadRequest, "invalid_id", "invalid uuid")
        }
        items, err := rm.Usage().GetByConversationID(c.Context(), id)
        if err != nil {
            return apiError(c, fiber.StatusInternalServerError, "query_failed", err.Error())
        }
        return c.JSON(fiber.Map{"data": items})
    }
}

// functionsListHandler returns available standard functions and arg constraints
func functionsListHandler(registry *functions.Registry) fiber.Handler {
    return func(c *fiber.Ctx) error {
        names := make([]fiber.Map, 0)
        for name := range registry.GetFunctions() {
            min, _ := registry.GetMinArgs(name)
            max, _ := registry.GetMaxArgs(name)
            names = append(names, fiber.Map{"name": name, "min_args": min, "max_args": max})
        }
        return c.JSON(fiber.Map{"data": names})
    }
}

type executeRequest struct {
    Args []string `json:"args"`
}

// functionExecuteHandler executes a standard function call
func functionExecuteHandler(registry *functions.Registry) fiber.Handler {
    return func(c *fiber.Ctx) error {
        name := c.Params("name")
        var req executeRequest
        if err := c.BodyParser(&req); err != nil {
            return apiError(c, fiber.StatusBadRequest, "bad_request", fmt.Sprintf("invalid body: %v", err))
        }
        res, err := registry.Call(name, req.Args)
        if err != nil {
            return apiError(c, fiber.StatusBadRequest, "function_error", err.Error())
        }
        return c.JSON(fiber.Map{"result": res})
    }
}

// apiError provides consistent error payloads
func apiError(c *fiber.Ctx, status int, code, msg string) error {
    return c.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": msg}})
}