package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `json:"username" validate:"required"`
	Email    string             `json:"email" validate:"required"`
	Password string             `json:"password" validate:"required"`
}
