package utils

import (
	"context"
	"satusehat-golang/models"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

func LogAudit(db *mongo.Database, log models.AuditLog) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	log.Timestamp = time.Now().Unix() // if using timestamp
	_, err := db.Collection("audit_logs").InsertOne(ctx, log)
	return err
}
