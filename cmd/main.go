package main

import (
	"fyp-api-gateway/apis"
	"log"
)

func main() {
	err := apis.CreateAPI()
	if err != nil {
		log.Fatal("(-) could not start microservices api! ", err)
	}
}
