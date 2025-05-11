package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IProducts struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Title       string             `bson:"title" json:"title"`
	Price       float64            `bson:"price" json:"price"`
	Description string             `bson:"description" json:"description"`
	ImageUrl    string             `bson:"image_url" json:"image_url"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	Rating      float64            `bson:"rating" json:"rating"`
	Comment     []ICommentData     `bson:"comment" json:"comment"`
}

type ICreateProduct struct {
	Title       string  `form:"title" binding:"required"`       // Cambiado a form:"title"
	Price       float64 `form:"price" binding:"required"`       // Cambiado a form:"price"
	Description string  `form:"description" binding:"required"` // Cambiado a form:"description"
	ImageUrl    string  `form:"image_url"`
}

type ICommentData struct {
	Comment   string    `bson:"comment" json:"comment" binding:"required"`
	Rating    int64     `bson:"rating" json:"rating" binding:"required"`
	Username  string    `bson:"username" json:"username" binding:"required"`
	Email     string    `bson:"email" json:"email" binding:"required"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}

type IComment struct {
	Comment string `form:"comment" binding:"required"`
	Rating  int64  `form:"rating" binding:"required"`
}
