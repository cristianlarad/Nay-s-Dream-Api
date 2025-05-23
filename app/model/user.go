package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"password"`
	Roles    string             `bson:"roles" json:"roles"`
	Email    string             `bson:"email" json:"email"`
}
type ResponseUser struct {
	Id       primitive.ObjectID `bson:"_id" json:"id"`
	Username string             `bson:"username" json:"username"`
	Roles    string             `bson:"roles" json:"roles"`
	Password string             `bson:"password"`

	Email string `bson:"email" json:"email"`
}
