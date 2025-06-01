package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Credential struct {
	ClientID     string `bson:"client_id"`
	ClientSecret string `bson:"client_secret"`
	TokenURL     string `bson:"token_url"`
}

type Token struct {
	AccessToken string    `bson:"access_token"`
	Expiry      time.Time `bson:"expiry"`
}

func GetValidToken(db *mongo.Database) (string, error) {
	var token Token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Find the current token
	err := db.Collection("tokens").FindOne(ctx, bson.M{}).Decode(&token)
	if err != nil && err != mongo.ErrNoDocuments {
		return "", err
	}

	// Check if expired or not found
	if token.AccessToken == "" || time.Now().After(token.Expiry.Add(-10*time.Second)) {
		// Get a new token
		newToken, expiry, err := GenerateNewToken(db) // Implement this function!
		if err != nil {
			return "", err
		}

		// Save to MongoDB
		token = Token{AccessToken: newToken, Expiry: expiry}
		_, err = db.Collection("tokens").ReplaceOne(ctx, bson.M{}, token, options.Replace().SetUpsert(true))
		if err != nil {
			return "", err
		}
	}

	return token.AccessToken, nil
}

func GenerateNewToken(db *mongo.Database) (string, time.Time, error) {
	// Get client_id, client_secret from MongoDB (credentials)
	var cred Credential
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := db.Collection("credentials").FindOne(ctx, bson.M{}).Decode(&cred)
	if err != nil {
		return "", time.Time{}, err
	}

	// Prepare POST request to token_url
	form := url.Values{}
	form.Add("client_id", cred.ClientID)
	form.Add("client_secret", cred.ClientSecret)
	form.Add("grant_type", "client_credentials")
	req, _ := http.NewRequest("POST", cred.TokenURL, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Make the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	// Calculate expiry
	expiry := time.Now().Add(time.Duration(result.ExpiresIn) * time.Second)
	return result.AccessToken, expiry, nil
}
