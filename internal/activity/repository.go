package activity

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Repository struct {
    collection *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
    coll := db.Collection("activities")

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    coll.Indexes().CreateOne(ctx, mongo.IndexModel{
        Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "triggered_at", Value: -1}},
    })

    return &Repository{collection: coll}
}

func (r *Repository) Create(ctx context.Context, a *Activity) error {
    a.ID = primitive.NewObjectID()
    a.TriggeredAt = time.Now()
    _, err := r.collection.InsertOne(ctx, a)
    return err
}

func (r *Repository) FindByUser(ctx context.Context, userID primitive.ObjectID, limit int64) ([]Activity, error) {
    opts := options.Find().
        SetSort(bson.D{{Key: "triggered_at", Value: -1}}).
        SetLimit(limit)

    cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var activities []Activity
    if err := cursor.All(ctx, &activities); err != nil {
        return nil, err
    }
    if activities == nil {
        activities = []Activity{}
    }
    return activities, nil
}