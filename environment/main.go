package environment

import (
	"os"

	"github.com/joho/godotenv"
)

var (
	DbHost     = ""
	DbPort     = ""
	DbUser     = ""
	DbName     = ""
	DbPassword = ""
	DbDriver   = ""
)

func InitEnv() {
	godotenv.Load()
	DbHost = GetEnv("DB_HOST")
	DbPort = GetEnv("DB_PORT")
	DbUser = GetEnv("DB_USER")
	DbName = GetEnv("DB_NAME")
	DbPassword = GetEnv("DB_PASSWORD")
	DbDriver = GetEnv("DB_DRIVER")
}

func GetEnv(key string) string {

	value := os.Getenv(key)

	if value == "" {
		panic("env variable " + key + " not found")
	}

	return value
}
