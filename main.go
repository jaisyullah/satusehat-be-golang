package main

import (
	"context"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"satusehat-golang/handlers"
)

func main() {
	e := echo.New()

	// Inisialisasi MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("satusehat_mirror")

	// Routing
	// resource: Encounter
	e.GET("/simrs/v1/encounter/:id", handlers.GetEncounter(db))
	e.POST("/simrs/v1/encounter/create", handlers.CreateEncounter(db))
	e.POST("/simrs/v1/encounter/update/:id", handlers.UpdateEncounter(db))
	e.PATCH("/simrs/v1/encounter/patch/:id", handlers.PatchEncounter(db))

	// resource: Location
	e.GET("/simrs/v1/location/:id", handlers.GetEncounter(db))
	e.POST("/simrs/v1/location/create", handlers.CreateLocation(db))
	e.POST("/simrs/v1/location/update/:id", handlers.UpdateLocation(db))
	e.PATCH("/simrs/v1/location/patch/:id", handlers.PatchLocation(db))

	// Get patient & practitioner
	e.GET("/simrs/v1/patient/:id", handlers.GetPatient(db))
	e.GET("/simrs/v1/practitioner/:id", handlers.GetPractitioner(db))

	// Credential endpoints
	e.GET("/simrs/v1/credentials", handlers.ListCredential(db))
	e.POST("/simrs/v1/credentials", handlers.InsertCredential(db))
	e.DELETE("/simrs/v1/credentials", handlers.DeleteCredential(db))

	//audit log
	e.GET("/simrs/v1/audit-logs", handlers.ListAuditLogs(db))

	e.Logger.Fatal(e.Start(":8080"))
}
