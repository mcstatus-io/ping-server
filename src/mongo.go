package main

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	CollectionApplications string = "applications"
	CollectionTokens       string = "tokens"
	CollectionRequestLog   string = "request_log"

	ErrMongoNotConnected error = errors.New("cannot use method as MongoDB is not connected")
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type Application struct {
	ID               string    `bson:"_id" json:"id"`
	Name             string    `bson:"name" json:"name"`
	ShortDescription string    `bson:"shortDescription" json:"shortDescription"`
	User             string    `bson:"user" json:"user"`
	Token            string    `bson:"token" json:"token"`
	RequestCount     uint64    `bson:"requestCount" json:"requestCount"`
	CreatedAt        time.Time `bson:"createdAt" json:"createdAt"`
}

type Token struct {
	ID           string    `bson:"_id" json:"id"`
	Name         string    `bson:"name" json:"name"`
	Token        string    `bson:"token" json:"token"`
	RequestCount uint64    `bson:"requestCount" json:"requestCount"`
	Application  string    `bson:"application" json:"application"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	LastUsedAt   time.Time `bson:"lastUsedAt" json:"lastUsedAt"`
}

func (c *MongoDB) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	parsedURI, err := url.Parse(*config.MongoDB)

	if err != nil {
		return err
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(*config.MongoDB))

	if err != nil {
		return err
	}

	c.Client = client
	c.Database = client.Database(strings.TrimPrefix(parsedURI.Path, "/"))

	return nil
}

func (c *MongoDB) GetTokenByToken(token string) (*Token, error) {
	if c.Client == nil {
		return nil, ErrMongoNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := c.Database.Collection(CollectionTokens).FindOne(ctx, bson.M{"token": token})

	if err := cur.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	var result Token

	if err := cur.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MongoDB) GetApplicationByID(id string) (*Application, error) {
	if c.Client == nil {
		return nil, ErrMongoNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	cur := c.Database.Collection(CollectionApplications).FindOne(ctx, bson.M{"_id": id})

	if err := cur.Err(); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, err
	}

	var result Application

	if err := cur.Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *MongoDB) UpdateToken(id string, update bson.M) error {
	if c.Client == nil {
		return ErrMongoNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := c.Database.Collection(CollectionTokens).UpdateOne(
		ctx,
		bson.M{"_id": id},
		update,
	)

	return err
}

func (c *MongoDB) IncrementApplicationRequestCount(id string) error {
	if c.Client == nil {
		return ErrMongoNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := c.Database.Collection(CollectionApplications).UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{
				"totalRequests": 1,
			},
		},
	)

	return err
}

func (c *MongoDB) UpsertRequestLog(query, update bson.M) error {
	if c.Client == nil {
		return ErrMongoNotConnected
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	_, err := c.Database.Collection(CollectionRequestLog).UpdateOne(
		ctx,
		query,
		update,
		&options.UpdateOptions{
			Upsert: PointerOf(true),
		},
	)

	return err
}

func (c *MongoDB) Close() error {
	if c.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

	defer cancel()

	return c.Client.Disconnect(ctx)
}
