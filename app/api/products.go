package api

import (
	"mgo-gin/app/cloudinary"
	"mgo-gin/app/model"
	"mgo-gin/app/repository"
	"mgo-gin/db"
	"mgo-gin/middlewares"
	err2 "mgo-gin/utils/err"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ApplyProductsAPI(app *gin.RouterGroup, resource *db.Resource, cldService *cloudinary.CloudinaryService) {
	productEntity := repository.NewProductEntity(resource, cldService)
	productRoute := app.Group("/product")

	productRoute.GET("", getAllProduct(productEntity))
	productRoute.POST("", createProduct(productEntity))
	productRoute.GET("/:productid", getOneProduct(productEntity))
	productRoute.POST("/:productid", updateProduct(productEntity))
	productRoute.POST("/:productid/add-comment", middlewares.AuthRequired(), addComment(productEntity))

}

func addComment(productEntity repository.IProduct) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		productid := ctx.Param("productid")
		username, exists := ctx.Get("username")
		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Nombre de usuario no encontrado en el contexto"})
			return
		}
		email, exists := ctx.Get("email")
		if !exists {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Email no encontrado en el contexto"})
			return
		}

		// Asegurarse de que son strings
		usernameStr, ok := username.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de nombre de usuario inválido en el contexto"})
			return
		}
		emailStr, ok := email.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Formato de email inválido en el contexto"})
			return
		}
		var comment model.IComment
		if err := ctx.ShouldBind(&comment); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos de comentario inválidos: " + err.Error()})
			return
		}
		commentedProduct, statusCode, err := productEntity.AddComment(productid, usernameStr, emailStr, comment)
		if err != nil {
			ctx.JSON(statusCode, gin.H{"error": err2.GetErrorMessage(err)})
			return
		}
		ctx.JSON(statusCode, gin.H{
			"message": "Comentario agregado exitosamente",
			"product": commentedProduct,
		})
	}
}

func updateProduct(productEntity repository.IProduct) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		productid := ctx.Param("productid")
		var productData model.ICreateProduct
		if err := ctx.ShouldBind(&productData); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos de producto inválidos: " + err.Error()})
			return
		}
		updatedProduct, statusCode, err := productEntity.UpdateProduct(productData, productid)
		if err != nil {
			ctx.JSON(statusCode, gin.H{"error": err2.GetErrorMessage(err)})
			return
		}
		ctx.JSON(statusCode, gin.H{
			"message": "Producto actualizado exitosamente",
			"product": updatedProduct,
		})
	}
}

func getOneProduct(productEntity repository.IProduct) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		productid := ctx.Param("productid")
		product, statusCode, err := productEntity.GetOneProduct(productid)
		response := map[string]interface{}{
			"product": product,
			"err":     err2.GetErrorMessage(err),
		}
		ctx.JSON(statusCode, response)
	}
}

func getAllProduct(productEntity repository.IProduct) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		pageStr := ctx.DefaultQuery("page", "1")
		perPageStr := ctx.DefaultQuery("perPage", "10") // Default to 10 items per page, adjust as needed
		search := ctx.DefaultQuery("search", "")
		maxPriceStr := ctx.DefaultQuery("maxPrice", "0")
		minPriceStr := ctx.DefaultQuery("minPrice", "0")
		page, err := strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1 // Default to page 1 if invalid
		}
		maxPrice, err := strconv.ParseFloat(maxPriceStr, 64) // Convert to float64
		if err != nil {
			// Handle error appropriately, e.g., return bad request or use a default
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid maxPrice format"})
			return
		}

		minPrice, err := strconv.ParseFloat(minPriceStr, 64) // Convert to float64
		if err != nil {
			// Handle error appropriately
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid minPrice format"})
			return
		}
		perPage, err := strconv.Atoi(perPageStr)
		if err != nil || perPage < 1 {
			perPage = 10 // Default to 10 per page if invalid
		}

		list, totalCount, statusCode, err := productEntity.GetAll(page, perPage, search, maxPrice, minPrice)

		var errorMsg interface{} // Use interface{} to be compatible with err2.GetErrorMessage or string
		if err != nil {
			// If you have a custom error handling package like `err2` as in your snippet:
			// errorMsg = err2.GetErrorMessage(err)
			// Otherwise, for a simple string error:
			errorMsg = err.Error()
		}

		// Calculate total pages
		var totalPages int64
		if perPage > 0 && totalCount > 0 { // Avoid division by zero and ensure totalCount is positive
			totalPages = (totalCount + int64(perPage) - 1) / int64(perPage)
		} else if totalCount == 0 {
			totalPages = 0 // No items, so 0 pages or 1 page depending on preference
		}

		response := gin.H{
			"products":   list,
			"total":      totalCount,
			"page":       page,
			"perPage":    perPage,
			"totalPages": totalPages,
			"error":      errorMsg,
		}
		ctx.JSON(statusCode, response)
	}
}

func createProduct(productEntity repository.IProduct) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		var productData model.ICreateProduct

		if err := ctx.ShouldBind(&productData); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Datos de producto inválidos: " + err.Error()})
			return
		}

		// Obtener el archivo de imagen del formulario
		file, header, err := ctx.Request.FormFile("image") // "image" es el nombre del campo en el formulario
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Se requiere un archivo de imagen: " + err.Error()})
			return
		}
		defer file.Close()

		// Llamar al método del repositorio para crear el producto y subir la imagen
		createdProduct, statusCode, err := productEntity.CreateOne(productData, file, header.Filename)
		if err != nil {
			// Usa err2.GetErrorMessage si lo tienes y es apropiado aquí
			ctx.JSON(statusCode, gin.H{"error": err2.GetErrorMessage(err)})
			return
		}

		ctx.JSON(statusCode, gin.H{
			"message": "Producto creado exitosamente",
			"product": createdProduct,
		})
	}
}
