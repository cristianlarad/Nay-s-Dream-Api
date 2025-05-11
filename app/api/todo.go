package api

import (
	"mgo-gin/app/form"
	"mgo-gin/app/model"
	"mgo-gin/app/repository"
	"mgo-gin/db"
	err2 "mgo-gin/utils/err"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ApplyToDoAPI(app *gin.RouterGroup, resource *db.Resource) {
	toDoEntity := repository.NewToDoEntity(resource)
	toDoRoute := app.Group("/todo")

	toDoRoute.GET("", getAllToDo(toDoEntity))
	toDoRoute.GET("/:id", getToDoById(toDoEntity))
	toDoRoute.POST("", createToDo(toDoEntity))
	toDoRoute.PUT("/:id", updateToDo(toDoEntity))

}

// GetAllToDo godoc
// @Summary Get all todos
// @Description Get all todo items
// @Tags todo
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} model.ToDo
// @Router /todo [get]
func getAllToDo(toDoEntity repository.IToDo) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		list, code, err := toDoEntity.GetAll()
		response := map[string]interface{}{
			"todo": list,
			"err":  err2.GetErrorMessage(err),
		}
		ctx.JSON(code, response)
	}
}

// CreateToDo godoc
// @Summary Create a new todo
// @Description Create a new todo item
// @Tags todo
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param todo body model.ICreateToDo true "Todo object"
// @Success 200 {object} model.ICreateToDo
// @Router /todo [post]
func createToDo(toDoEntity repository.IToDo) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {

		todoReq := model.ICreateToDo{}
		if err := ctx.BindJSON(&todoReq); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		todo, code, err := toDoEntity.CreateOne(todoReq)
		response := map[string]interface{}{
			"todo": todo,
			"err":  err2.GetErrorMessage(err),
		}
		ctx.JSON(code, response)
	}
}

// GetToDoById godoc
// @Summary Get todo by ID
// @Description Get a todo item by its ID
// @Tags todo
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Todo ID"
// @Success 200 {object} model.ToDo
// @Router /todo/{id} [get]
func getToDoById(toDoEntity repository.IToDo) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		todo, code, err := toDoEntity.GetOneByID(id)
		response := map[string]interface{}{
			"todo": todo,
			"err":  err2.GetErrorMessage(err),
		}
		ctx.JSON(code, response)
	}
}

// UpdateToDo godoc
// @Summary Update a todo
// @Description Update a todo item by its ID
// @Tags todo
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Todo ID"
// @Param todo body form.ToDoForm true "Todo update object"
// @Success 200 {object} model.ToDo
// @Router /todo/{id} [put]
func updateToDo(toDoEntity repository.IToDo) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		id := ctx.Param("id")
		todoReq := form.ToDoForm{}
		if err := ctx.Bind(&todoReq); err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
		todo, code, err := toDoEntity.Update(id, todoReq)
		response := map[string]interface{}{
			"todo": todo,
			"err":  err2.GetErrorMessage(err),
		}
		ctx.JSON(code, response)
	}
}
