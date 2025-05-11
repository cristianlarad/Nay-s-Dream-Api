package repository

import (
	"mgo-gin/app/form"
	"mgo-gin/app/model"
	"mgo-gin/db"
	"mgo-gin/utils"
	"net/http"

	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ToDoEntity IToDo

type toDoEntity struct {
	resource *db.Resource
	repo     *mongo.Collection
}

type IToDo interface {
	GetAll() ([]model.ICreateToDo, int, error)
	CreateOne(todoForm model.ICreateToDo) (model.ICreateToDo, int, error)
	GetOneByID(id string) (*model.ToDo, int, error) // need return pointer
	Update(id string, todo form.ToDoForm) (model.ToDo, int, error)
}

// func NewToDoEntity
func NewToDoEntity(resource *db.Resource) IToDo {
	toDoRepo := resource.DB.Collection("todo")
	ToDoEntity = &toDoEntity{resource: resource, repo: toDoRepo}
	return ToDoEntity
}

func (entity *toDoEntity) GetAll() ([]model.ICreateToDo, int, error) {
	toDoList := []model.ICreateToDo{}
	ctx, cancel := initContext()
	defer cancel()
	cursor, err := entity.repo.Find(ctx, bson.M{})

	if err != nil {
		return []model.ICreateToDo{}, 400, err
	}

	for cursor.Next(ctx) {
		var todo model.ICreateToDo
		err = cursor.Decode(&todo)
		if err != nil {
			logrus.Print(err)
		}
		toDoList = append(toDoList, todo)
	}
	return toDoList, http.StatusOK, nil
}

func (entity *toDoEntity) CreateOne(todoForm model.ICreateToDo) (model.ICreateToDo, int, error) {
	todo := model.ICreateToDo{
		Id:          primitive.NewObjectID(),
		Name:        todoForm.Name,
		Priority:    todoForm.Priority,
		Status:      utils.PENDDING,
		Description: todoForm.Description,
	}

	ctx, cancel := initContext()
	defer cancel()

	_, err := entity.repo.InsertOne(ctx, todo)
	if err != nil {
		return model.ICreateToDo{}, 400, err
	}

	return todo, http.StatusOK, nil
}

func (entity *toDoEntity) GetOneByID(id string) (*model.ToDo, int, error) {
	var todo model.ToDo
	ctx, cancel := initContext()
	defer cancel()
	logrus.Print(id)
	objID, _ := primitive.ObjectIDFromHex(id)

	err := entity.repo.FindOne(ctx, bson.M{"_id": objID}).Decode(&todo)
	if err != nil {
		return nil, http.StatusNotFound, err
	}

	return &todo, http.StatusOK, nil
}

func (entity *toDoEntity) Update(id string, todoForm form.ToDoForm) (model.ToDo, int, error) {
	var todo *model.ToDo
	ctx, cancel := initContext()

	defer cancel()
	objID, _ := primitive.ObjectIDFromHex(id)

	todo, _, err := entity.GetOneByID(id)
	if err != nil {
		return model.ToDo{}, http.StatusNotFound, nil
	}

	err = copier.Copy(todo, todoForm) // this is why we need return a pointer: to copy value
	if err != nil {
		logrus.Error(err)
		return model.ToDo{}, getHTTPCode(err), err
	}

	isReturnNewDoc := options.After
	opts := &options.FindOneAndUpdateOptions{
		ReturnDocument: &isReturnNewDoc,
	}
	err = entity.repo.FindOneAndUpdate(ctx, bson.M{"_id": objID}, bson.M{"$set": todo}, opts).Decode(&todo)

	return *todo, http.StatusOK, nil
}
