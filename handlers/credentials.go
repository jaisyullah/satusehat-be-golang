package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Credential struct {
	ClientID     string `json:"client_id" bson:"client_id"`
	ClientSecret string `json:"client_secret" bson:"client_secret"`
	TokenURL     string `json:"token_url" bson:"token_url"`
}

func DeleteCredential(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := db.Collection("credentials").DeleteMany(ctx, bson.M{})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to delete credentials: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
	}
}

// ListCredential : get current credential
func ListCredential(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cred Credential

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := db.Collection("credentials").FindOne(ctx, bson.M{}).Decode(&cred)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.JSON(http.StatusOK, nil) // Kalau tidak ada dokumen, kirimkan `null`
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to get credential: %v", err),
			})
		}

		return c.JSON(http.StatusOK, cred)
	}
}

func InsertCredential(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		var cred Credential
		if err := c.Bind(&cred); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		opts := options.Replace().SetUpsert(true)
		_, err := db.Collection("credentials").ReplaceOne(ctx, bson.M{}, cred, opts)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": fmt.Sprintf("Failed to insert/update credential: %v", err),
			})
		}

		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}
