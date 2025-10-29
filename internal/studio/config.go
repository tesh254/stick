package studio

import (
    "os"
)

type Config struct {
    Port           string
    AllowedOrigins string
    Env            string
}

func LoadConfig() Config {
    port := getenvDefault("STICK_STUDIO_PORT", "8080")
    origins := getenvDefault("STICK_STUDIO_ALLOWED_ORIGINS", "*")
    env := getenvDefault("STICK_ENV", "development")
    return Config{
        Port:           port,
        AllowedOrigins: origins,
        Env:            env,
    }
}

func getenvDefault(key, def string) string {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    return v
}