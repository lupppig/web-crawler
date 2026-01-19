package storage

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

const (
	DBName             = "crawledContent"
	CollectionContent  = "content"
	CollectionMetaData = "url_metadata"
)

type CrawlContent struct {
	Title   string    `bson:"title"`
	Body    string    `bson:"body"`
	Path    string    `bson:"path"`
	AddedAt time.Time `bson:"added_at"`
}

type Storage struct {
	client *mongo.Client
	dbName string
}

func NewStorage(dbName string) *Storage {
	if dbName == "" {
		dbName = DBName
	}
	return &Storage{
		dbName: dbName,
	}
}

func (s *Storage) Connect(connStr string) error {
	client, err := mongo.Connect(options.Client().ApplyURI(connStr))
	if err != nil {
		return fmt.Errorf("could not connect to mongdb: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return fmt.Errorf("could not ping: %w", err)
	}
	s.client = client

	s.ensureIndexes(ctx)

	fmt.Println("database connected successfully")
	return nil
}

func (s *Storage) ensureIndexes(ctx context.Context) {
	collection := s.client.Database(s.dbName).Collection(CollectionContent)
	textModelIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "body", Value: "text"},
		},
	}
	if _, err := collection.Indexes().CreateOne(ctx, textModelIndex); err != nil {
		fmt.Println("unable to create text index on collection: " + err.Error())
	}

	collection = s.client.Database(s.dbName).Collection(CollectionMetaData)

	uniqueModelIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "path", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := collection.Indexes().CreateOne(ctx, uniqueModelIndex); err != nil {
		fmt.Println("unable to create unique index on url: " + err.Error())
	}

	uniqueModelIndexUrl := mongo.IndexModel{
		Keys: bson.D{
			{Key: "url", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	if _, err := collection.Indexes().CreateOne(ctx, uniqueModelIndexUrl); err != nil {
		fmt.Println("unable to create unique index on url: " + err.Error())
	}
}

func (s *Storage) AddContent(ctx context.Context, content CrawlContent) error {
	col := s.client.Database(s.dbName).Collection(CollectionContent)
	_, err := col.InsertOne(ctx, content)
	if err != nil {
		return fmt.Errorf("failed to insert data in mongo: %w", err)
	}
	fmt.Printf("Successfully saved to DB: %s (%s)\n", content.Path, content.Title)
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	if s.client != nil {
		return s.client.Disconnect(ctx)
	}
	return nil
}
