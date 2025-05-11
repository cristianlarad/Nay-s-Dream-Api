package repository

import (
	"bytes"
	"context"
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"math"
	"mgo-gin/app/cloudinary"
	"mgo-gin/app/model"
	"mgo-gin/db"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type productEntity struct {
	resource          *db.Resource
	repo              *mongo.Collection
	cloudinaryService *cloudinary.CloudinaryService // Added CloudinaryService
}

type IProduct interface {
	GetAll(page, perPage int, search string, maxPrice float64, minPrice float64) (products []model.IProducts, totalCount int64, statusCode int, err error)
	GetOneProduct(productid string) (product model.IProducts, statusCode int, err error)
	CreateOne(productData model.ICreateProduct, imageFile multipart.File, imageFilename string) (model.IProducts, int, error)
	UpdateProduct(productData model.ICreateProduct, productid string) (model.IProducts, int, error)
	AddComment(productid string, username string, email string, comment model.IComment) (model.IProducts, int, error)
}

// Updated NewProductEntity to accept CloudinaryService and return concrete type
func NewProductEntity(resource *db.Resource, cldService *cloudinary.CloudinaryService) *productEntity {
	productRepo := resource.DB.Collection("product")
	return &productEntity{
		resource:          resource,
		repo:              productRepo,
		cloudinaryService: cldService, // Initialize CloudinaryService
	}
}

func (entity *productEntity) AddComment(productid string, username string, email string, comment model.IComment) (model.IProducts, int, error) {
	ctx, cancel := initContext()
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(productid)

	product := model.IProducts{}
	err := entity.repo.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		log.Printf("Error finding product: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	product.Comment = append(product.Comment, model.ICommentData{
		Comment:   comment.Comment,
		Rating:    comment.Rating,
		Username:  username,
		Email:     email,
		CreatedAt: time.Now().UTC(),
	})
	var TotalRating float64
	if len(product.Comment) > 0 {
		for _, comment := range product.Comment {
			TotalRating += float64(comment.Rating)
		}
		averageRating := TotalRating / float64(len(product.Comment))
		product.Rating = math.Round(averageRating*10) / 10
	} else {
		product.Rating = 1
	}
	updatePayload := bson.M{
		"$set": bson.M{
			"comment": product.Comment, // El campo BSON es "comment"
			"rating":  product.Rating,
		},
	}
	_, err = entity.repo.UpdateOne(ctx, bson.M{"_id": objID}, updatePayload)
	if err != nil {
		log.Printf("Error updating product: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	return product, http.StatusOK, nil
}

func (entity *productEntity) UpdateProduct(productData model.ICreateProduct, productid string) (model.IProducts, int, error) {
	ctx, cancel := initContext()
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(productid)

	product := model.IProducts{}
	err := entity.repo.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		log.Printf("Error finding product: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	product.Title = productData.Title
	product.Description = productData.Description
	product.Price = productData.Price
	_, err = entity.repo.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": product})
	if err != nil {
		log.Printf("Error updating product: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	return product, http.StatusOK, nil
}

func (entity *productEntity) GetOneProduct(productid string) (product model.IProducts, statusCode int, err error) {
	ctx, cancel := initContext()
	defer cancel()

	objID, _ := primitive.ObjectIDFromHex(productid)
	err = entity.repo.FindOne(ctx, bson.M{"_id": objID}).Decode(&product)
	if err != nil {
		log.Printf("Error finding product: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	return product, http.StatusOK, nil
}
func (entity *productEntity) CreateOne(productData model.ICreateProduct, imageFile multipart.File, imageFilename string) (model.IProducts, int, error) {

	img, _, err := image.Decode(imageFile)
	if err != nil {
		log.Printf("Error decodificando la imagen original: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	if img.Bounds().Dx() > 1200 { // Solo redimensionar si es más ancha de 1200px
		img = imaging.Resize(img, 1200, 0, imaging.Lanczos)
	}
	var webpBuffer bytes.Buffer

	if err := jpeg.Encode(&webpBuffer, img, nil); err != nil {
		log.Printf("Error codificando imagen a JPEG: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	originalExt := filepath.Ext(imageFilename)
	newImageFilename := strings.TrimSuffix(imageFilename, originalExt) + ".webp"
	log.Printf("Imagen procesada a WebP: %s, tamaño nuevo: %d bytes\n", newImageFilename, webpBuffer.Len())

	imageUrl, err := entity.cloudinaryService.UploadImage(&webpBuffer, newImageFilename)
	if err != nil {
		log.Printf("Error uploading image to Cloudinary: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}
	log.Printf("Image uploaded to Cloudinary: %s\n", imageUrl)

	// Step 2: Prepare product data for MongoDB
	newProduct := model.IProducts{
		Id:          primitive.NewObjectID(),
		Title:       productData.Title,
		Description: productData.Description,
		Price:       productData.Price,
		ImageUrl:    imageUrl,
		CreatedAt:   time.Now().UTC(),
		Rating:      1,
	}

	// Step 3: Insert into MongoDB
	// Using context.TODO() as a placeholder for your actual context logic
	ctx := context.TODO()
	_, err = entity.repo.InsertOne(ctx, newProduct)
	if err != nil {
		log.Printf("Error inserting product into MongoDB: %v\n", err)
		return model.IProducts{}, http.StatusInternalServerError, err
	}

	return newProduct, http.StatusCreated, nil
}

func (entity *productEntity) GetAll(page, perPage int, search string, maxPrice float64, minPrice float64) ([]model.IProducts, int64, int, error) {
	productList := []model.IProducts{}
	ctx := context.TODO() // Consider using a context propagated from the handler

	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 10 // Default items per page
	}
	skip := int64((page - 1) * perPage)

	findOptions := options.Find()
	findOptions.SetSkip(skip)
	findOptions.SetLimit(int64(perPage))
	// Example: findOptions.SetSort(bson.D{{"createdAt", -1}}) // Sort by creation_date descending

	filter := bson.M{}
	if search != "" {
		// Apply case-insensitive regex search on 'title'
		filter["title"] = bson.M{"$regex": search, "$options": "i"}
	}
	priceFilter := bson.M{}
	if minPrice > 0 {
		priceFilter["$gte"] = minPrice
	}
	if maxPrice > 0 {
		priceFilter["$lte"] = maxPrice
	}

	if len(priceFilter) > 0 {
		filter["price"] = priceFilter // Assuming your product model has a 'price' field
	}
	cursor, err := entity.repo.Find(ctx, filter, findOptions) // Use entity.repo and the filter
	if err != nil {
		log.Printf("Error finding products: %v\n", err)
		return []model.IProducts{}, 0, http.StatusInternalServerError, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &productList); err != nil {
		log.Printf("Error decoding products: %v\n", err)
		return []model.IProducts{}, 0, http.StatusInternalServerError, err
	}

	// Get total count of documents matching the filter for pagination
	totalCount, err := entity.repo.CountDocuments(ctx, filter) // Use entity.repo and the same filter
	if err != nil {
		log.Printf("Error counting documents: %v\n", err)
		return []model.IProducts{}, 0, http.StatusInternalServerError, err
	}

	return productList, totalCount, http.StatusOK, nil
}
