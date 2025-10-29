package studio

import (
    "github.com/tesh254/stick/internal/db"
    "github.com/tesh254/stick/internal/functions"
)

// RepoManagerFactory allows injecting repositories (real or mocks)
type RepoManagerFactory func() (db.RepositoryManager, error)

// FuncRegistryFactory allows injecting a functions registry
type FuncRegistryFactory func() *functions.Registry