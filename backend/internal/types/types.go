package types

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty"`
	Name     string             `json:"username" validate:"required"`
	Email    string             `json:"email" validate:"required"`
	Password string             `json:"password" validate:"required"`
}

type Book struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title" json:"title" validate:"required"`
	Author     string             `bson:"author" json:"author" validate:"required"`
	Genre      string             `bson:"genre" json:"genre" validate:"required"`
	CoverImage string             `bson:"coverimage" json:"coverimage" validate:"required"`
	File       string             `bson:"file" json:"file" validate:"required"`
	CreatedBy  string             `bson:"created_by" json:"created_by"`
	CreatedAt  string             `bson:"created_at" json:"created_at"`
	UpdatedAt  string             `bson:"updated_at" json:"updated_at"`
}
