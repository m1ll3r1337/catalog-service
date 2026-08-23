package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"

	"github.com/m1ll3r1337/catalog-service/internal/app/config/section"
)

type Config struct {
	Repository section.Repository `required:"true"`
	Monitor    section.Monitor    `required:"true"`
	Processor  section.Processor  `required:"true"`
}

var Root Config

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file: ", err)
	}

	err = envconfig.Process("APP", &Root)
	if err != nil {
		log.Fatal("Error processing environment variables: ", err)
	}
}
