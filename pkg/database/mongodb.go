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
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // increased from 10s
    defer cancel()

    // These options fix TLS issues with Atlas
    clientOptions := options.Client().
        ApplyURI(uri).
        SetServerSelectionTimeout(30 * time.Second).
        SetConnectTimeout(30 * time.Second)

    client, err := mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
    }

    // Ping to confirm connection
    pingCtx, pingCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer pingCancel()

    if err := client.Ping(pingCtx, nil); err != nil {
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
    m.Client.Disconnect(ctx)
}