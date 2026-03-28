package reminder

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
    coll := db.Collection("reminders")

    // Index on user_id for fast per-user queries
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    indexes := []mongo.IndexModel{
        {Keys: bson.D{{Key: "user_id", Value: 1}}},
        {Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}}},
    }
    coll.Indexes().CreateMany(ctx, indexes)

    return &Repository{collection: coll}
}

func (r *Repository) Create(ctx context.Context, reminder *Reminder) error {
    reminder.ID = primitive.NewObjectID()
    reminder.CreatedAt = time.Now()
    reminder.UpdatedAt = time.Now()
    reminder.Triggered = false
    _, err := r.collection.InsertOne(ctx, reminder)
    return err
}

func (r *Repository) FindAllByUser(ctx context.Context, userID primitive.ObjectID) ([]Reminder, error) {
    opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
    cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID}, opts)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var reminders []Reminder
    if err := cursor.All(ctx, &reminders); err != nil {
        return nil, err
    }
    // Return empty slice, never nil (cleaner JSON: [] not null)
    if reminders == nil {
        reminders = []Reminder{}
    }
    return reminders, nil
}

func (r *Repository) FindOne(ctx context.Context, id, userID primitive.ObjectID) (*Reminder, error) {
    var reminder Reminder
    // Always scope to userID — prevents user A from reading user B's reminders
    err := r.collection.FindOne(ctx, bson.M{
        "_id":     id,
        "user_id": userID,
    }).Decode(&reminder)
    if err != nil {
        return nil, err
    }
    return &reminder, nil
}

func (r *Repository) Update(ctx context.Context, id, userID primitive.ObjectID, fields bson.M) (*Reminder, error) {
    fields["updated_at"] = time.Now()

    after := options.After
    opts := options.FindOneAndUpdate().SetReturnDocument(after)

    var updated Reminder
    err := r.collection.FindOneAndUpdate(
        ctx,
        bson.M{"_id": id, "user_id": userID},
        bson.M{"$set": fields},
        opts,
    ).Decode(&updated)
    if err != nil {
        return nil, err
    }
    return &updated, nil
}

func (r *Repository) Delete(ctx context.Context, id, userID primitive.ObjectID) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{
        "_id":     id,
        "user_id": userID,
    })
    if err != nil {
        return err
    }
    if result.DeletedCount == 0 {
        return mongo.ErrNoDocuments
    }
    return nil
}

func (r *Repository) ToggleEnabled(ctx context.Context, id, userID primitive.ObjectID) (*Reminder, error) {
    // First fetch current state
    current, err := r.FindOne(ctx, id, userID)
    if err != nil {
        return nil, err
    }

    return r.Update(ctx, id, userID, bson.M{
        "enabled": !current.Enabled,
    })
}