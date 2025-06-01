package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"satusehat-golang/models"
	"satusehat-golang/utils"
)

func CreateLocation(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		body, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
		}
		c.Request().Body = io.NopCloser(bytes.NewReader(body))

		token, err := utils.GetValidToken(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get token"})
		}

		req, _ := http.NewRequest("POST", "https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Location", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to post to satusehat"})
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		var resourceID string
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			// Unmarshal response to extract ID
			var encResp models.LocationResponse
			if err := json.Unmarshal(respBody, &encResp); err == nil {
				resourceID = encResp.ID
				// Insert into MongoDB
				var doc interface{}
				if err := json.Unmarshal(respBody, &doc); err == nil {
					_, _ = db.Collection("locations").InsertOne(context.Background(), doc)
				}
			}

			// Save to audit log with ResourceID
			_ = utils.LogAudit(db, models.AuditLog{
				User:       "Admin", // Replace with JWT data if available
				Action:     "create",
				Resource:   "location",
				ResourceID: resourceID, // <-- use extracted ID here
				StatusCode: resp.StatusCode,
				Details: map[string]interface{}{
					"requestBody":  json.RawMessage(body),
					"responseBody": json.RawMessage(respBody),
				},
			})
		}

		return c.JSONBlob(resp.StatusCode, respBody)
	}
}

func UpdateLocation(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		locationID := c.Param("id")
		if locationID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing location ID"})
		}

		var location map[string]interface{}
		if err := c.Bind(&location); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		}

		reqBody, err := json.Marshal(location)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal JSON"})
		}

		token, err := utils.GetValidToken(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get token"})
		}

		url := fmt.Sprintf("https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Location/%s", locationID)

		req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(reqBody))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
		}

		req.Header.Set("Content-Type", "application/fhir+json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send request"})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated {
			// Save to audit log
			_ = utils.LogAudit(db, models.AuditLog{
				User:       "Admin",
				Action:     "put",
				Resource:   "location",
				ResourceID: locationID,
				StatusCode: resp.StatusCode,
				Details: map[string]interface{}{
					"requestBody":  json.RawMessage(reqBody),
					"responseBody": json.RawMessage(body),
				},
			})

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			filter := bson.M{"id": locationID}
			update := bson.M{"$set": location}
			_, err := db.Collection("locations").UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to mirror to MongoDB"})
			}
		}

		return c.Blob(resp.StatusCode, "application/fhir+json", body)
	}
}

func PatchLocation(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		locationID := c.Param("id")
		if locationID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing location ID"})
		}

		// Read JSON Patch operations
		var patchOps []map[string]interface{}
		if err := c.Bind(&patchOps); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid JSON body"})
		}

		reqBody, err := json.Marshal(patchOps)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to marshal JSON patch"})
		}

		// Get token
		token, err := utils.GetValidToken(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get token"})
		}

		// Build PATCH request
		url := fmt.Sprintf("https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Location/%s", locationID)
		req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(reqBody))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
		}

		req.Header.Set("Content-Type", "application/json-patch+json")
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send request"})
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		// Audit log (record both request and response bodies)
		_ = utils.LogAudit(db, models.AuditLog{
			User:       "Admin", // Extract from auth context if available
			Action:     "patch",
			Resource:   "location",
			ResourceID: locationID,
			StatusCode: resp.StatusCode,
			Details: map[string]interface{}{
				"requestBody":  json.RawMessage(reqBody),
				"responseBody": json.RawMessage(respBody),
			},
		})

		// If successful, mirror updated fields to MongoDB
		if resp.StatusCode == http.StatusOK {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			update := bson.M{}
			for _, op := range patchOps {
				if op["op"] == "replace" && op["path"] != nil && op["value"] != nil {
					path := strings.TrimPrefix(op["path"].(string), "/")
					update[path] = op["value"]
				}
			}
			if len(update) > 0 {
				filter := bson.M{"id": locationID}
				_, err := db.Collection("locations").UpdateOne(ctx, filter, bson.M{"$set": update})
				if err != nil {
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to mirror to MongoDB"})
				}
			}
		}

		// Return SatuSehat response
		return c.JSONBlob(resp.StatusCode, respBody)
	}
}

func GetLocation(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		locationID := c.Param("id")
		if locationID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing location ID"})
		}

		// Example: get token for SatuSehat API
		token, err := utils.GetValidToken(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get token"})
		}

		url := fmt.Sprintf("https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Location/%s", locationID)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create request"})
		}

		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send request"})
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)

		// Log audit for the GET request
		_ = utils.LogAudit(db, models.AuditLog{
			User:       "Admin", // Extract from JWT/auth context in real app
			Action:     "get",
			Resource:   "location",
			ResourceID: locationID,
			StatusCode: resp.StatusCode,
			Details: map[string]interface{}{
				"queryParams": c.QueryParams(),
				"response":    json.RawMessage(body),
			},
		})

		return c.JSONBlob(resp.StatusCode, body)
	}
}
