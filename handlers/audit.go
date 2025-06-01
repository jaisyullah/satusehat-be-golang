package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"satusehat-golang/models"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// decodeResponseField decodes the BSON binary response field from the audit log details
func decodeResponseField(log *models.AuditLog) error {
	rawBinary, ok := log.Details["response"].(primitive.Binary)
	if !ok {
		// response field is not binary, maybe already decoded or missing
		return nil
	}

	// decode JSON bytes
	var decoded interface{}
	if err := json.Unmarshal(rawBinary.Data, &decoded); err != nil {
		return err
	}

	// replace binary data with decoded JSON
	log.Details["response"] = decoded
	return nil
}

func ListAuditLogs(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		filter := bson.M{}
		resource := c.QueryParam("resource")
		if resource != "" {
			filter["resource"] = resource
		}

		cur, err := db.Collection("audit_logs").Find(ctx, filter)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch logs"})
		}
		defer cur.Close(ctx)

		var logs []models.AuditLog
		if err := cur.All(ctx, &logs); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to decode logs"})
		}

		// Decode each audit log response field from binary to JSON
		for i := range logs {
			if err := decodeResponseField(&logs[i]); err != nil {
				// optionally handle decode error, here we just continue
			}
		}

		return c.JSON(http.StatusOK, logs)
	}
}
