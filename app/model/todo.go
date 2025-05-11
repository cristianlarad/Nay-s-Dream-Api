package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type ToDo struct {
	Id   primitive.ObjectID `bson:"_id" json:"id"`
	Name string             `bson:"name" json:"name"`
}
type ICreateToDo struct {
	Id          primitive.ObjectID `bson:"_id" json:"id"`
	Name        string             `bson:"name" json:"name" binding:"required"`
	Priority    string             `bson:"priority" json:"priority" binding:"required"`
	Status      string             `bson:"status" json:"status"`
	Description string             `bson:"description" json:"description" binding:"required"`
}
