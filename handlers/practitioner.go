package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"satusehat-golang/models"
	"satusehat-golang/utils"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetPractitioner(db *mongo.Database) echo.HandlerFunc {
	return func(c echo.Context) error {
		PractitionerID := c.Param("id")
		if PractitionerID == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Missing Practitioner ID"})
		}

		// Example: get token for SatuSehat API
		token, err := utils.GetValidToken(db)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to get token"})
		}

		url := fmt.Sprintf("https://api-satusehat-stg.dto.kemkes.go.id/fhir-r4/v1/Practitioner/%s", PractitionerID)
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
			Resource:   "Practitioner",
			ResourceID: PractitionerID,
			StatusCode: resp.StatusCode,
			Details: map[string]interface{}{
				"queryParams": c.QueryParams(),
				"response":    json.RawMessage(body),
			},
		})

		return c.JSONBlob(resp.StatusCode, body)
	}
}
