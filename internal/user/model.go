package user

import (
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"      json:"id"`
    Name      string             `bson:"name"               json:"name"`
    Email     string             `bson:"email"              json:"email"`
    Password  string             `bson:"password"           json:"-"` 
    CreatedAt time.Time          `bson:"created_at"         json:"createdAt"`
    UpdatedAt time.Time          `bson:"updated_at"         json:"updatedAt"`
}

// Safe version to return to clients (no password)
type UserResponse struct {
    ID        string    `json:"id"`
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"createdAt"`
}

func (u *User) ToResponse() UserResponse {
    return UserResponse{
        ID:        u.ID.Hex(),
        Name:      u.Name,
        Email:     u.Email,
        CreatedAt: u.CreatedAt,
    }
}