package app

import (
	"mgo-gin/app/api"
	"mgo-gin/app/cloudinary"
	"mgo-gin/db"
	"mgo-gin/middlewares"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Routes struct {
}

func (app Routes) StartGin() {
	r := gin.Default()
	// Configure CORS
	r.Use(gin.Logger())
	r.Use(middlewares.NewRecovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           50 * time.Second,
	}))
	r.GET("swagger/*any", middlewares.NewSwagger())

	publicRoute := r.Group("/api/v1")
	resource, err := db.InitResource()
	if err != nil {
		logrus.Error(err)
	}
	defer resource.Close()
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
	//r.Static("/template/css", "./template/css")
	//r.Static("/template/images", "./template/images")
	r.Static("/template", "./template")
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		logrus.Fatal("Cloudinary environment variables not set (CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET)")
	}

	cldService, err := cloudinary.NewCloudinaryService(cloudName, apiKey, apiSecret)
	if err != nil {
		logrus.Fatalf("Failed to initialize Cloudinary service: %v", err)
	}
	// Rutas públicas (sin autenticación)
	api.ApplyUserAPI(publicRoute, resource)
	api.ApplyProductsAPI(publicRoute, resource, cldService)
	// Rutas protegidas (con autenticación)
	protectedRoute := publicRoute.Group("")
	protectedRoute.Use(middlewares.AuthRequired())

	r.NoRoute(func(context *gin.Context) {
		context.File("./template/index.html")
	})

	r.Run(":8080")
}
