package database

import (
    "context"
    "log"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
    Client *mongo.Client
    DB     *mongo.Database
}

func NewMongoDB(uri, dbName string) *MongoDB {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI(uri)
    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
    }

    // Ping to confirm connection
    if err := client.Ping(ctx, nil); err != nil {
        log.Fatalf("❌ MongoDB ping failed: %v", err)
    }

    log.Println("✅ Connected to MongoDB")

    return &MongoDB{
        Client: client,
        DB:     client.Database(dbName),
    }
}

func (m *MongoDB) Disconnect() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := m.Client.Disconnect(ctx); err != nil {
        log.Printf("Error disconnecting MongoDB: %v", err)
    }
}