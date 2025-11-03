package main

import (
	"flag"
	"log"

	"api/src/config"
	"api/src/tools"
)

func main() {
	var generateModels bool
	flag.BoolVar(&generateModels, "generate-models", false, "Generate models from existing postgres database")
	flag.Parse()

	// Connect to database
	config.ConnectToDatabase()

	if generateModels {
		log.Println("[NOTICE] Generating models from database...")
		if err := tools.GenerateModelsFromDatabase(); err != nil {
			log.Fatal("[ERROR] Failed to generate models:", err)
		}
		log.Println("[NOTICE] Model generation completed!")
		return
	}
}
