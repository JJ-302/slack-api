package db

import (
	"context"
	"log"
	"projects/slack-api/config"

	"cloud.google.com/go/datastore"
)

type Token struct {
	Token string
}

func New(token string) *Token {
	return &Token{
		Token: token,
	}
}

func (token *Token) Save(uid string) error {
	ctx := context.Background()
	projectID := config.Config.ProjectID
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to create client: %v", err)
		return err
	}

	kind := "Token"
	tokenKey := datastore.NameKey(kind, uid, nil)
	_, err = client.Put(ctx, tokenKey, token)
	return err
}

func (token *Token) Get(uid string) *Token {
	ctx := context.Background()
	projectID := config.Config.ProjectID
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("Failed to register token: %v", err)
		return token
	}

	kind := "Token"
	tokenKey := datastore.NameKey(kind, uid, nil)
	if err = client.Get(ctx, tokenKey, token); err != nil {
		log.Printf("Failed to get token: %v", err)
		return token
	}
	return token
}
