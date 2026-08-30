package user

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
    coll := db.Collection("users")

    // Create unique index on email
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    indexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    coll.Indexes().CreateOne(ctx, indexModel)

    return &Repository{collection: coll}
}

func (r *Repository) Create(ctx context.Context, user *User) error {
    user.ID = primitive.NewObjectID()
    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()
    _, err := r.collection.InsertOne(ctx, user)
    return err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id primitive.ObjectID) (*User, error) {
    var user User
    err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Update writes only the named fields, mirroring the reminder repository. It
// previously $set the whole User struct, which meant every caller had to hold a
// fully-populated record or silently blank the fields it had not loaded, and
// sent immutable _id and created_at on every write.
func (r *Repository) Update(ctx context.Context, id primitive.ObjectID, fields bson.M) error {
    if len(fields) == 0 {
        return nil
    }
    fields["updated_at"] = time.Now()

    result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": fields})
    if err != nil {
        return err
    }
    if result.MatchedCount == 0 {
        return mongo.ErrNoDocuments
    }
    return nil
}

// UpdatePassword stores an already-hashed password. Hashing stays in the
// service layer; the repository never sees a plaintext password.
func (r *Repository) UpdatePassword(ctx context.Context, id primitive.ObjectID, hashedPassword string) error {
    return r.Update(ctx, id, bson.M{"password": hashedPassword})
}

func (r *Repository) Delete(ctx context.Context, id primitive.ObjectID) error {
    result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    if err != nil {
        return err
    }
    if result.DeletedCount == 0 {
        return mongo.ErrNoDocuments
    }
    return nil
}
