package db

// This file documents the database schema and migration functions for Turso database

// Schema Version: 1
// 
// Tables:
// 1. conversations - Stores conversation metadata
// 2. messages - Stores individual messages within conversations
// 3. usage - Stores token usage information for conversations
//
// Relations:
// - messages.conversation_id -> conversations.id (FK)
// - usage.conversation_id -> conversations.id (FK)

// Migration functions would go here if we needed to update the schema over time
// For now, the createTables function in db.go serves as our initial migration

// Versioned migration example (if needed in the future):
// func migrateToVersion1(db *sql.DB) error {
//     // Migration logic for version 1
//     return nil
// }
// 
// func migrateToVersion2(db *sql.DB) error {
//     // Migration logic for version 2
//     return nil
// }