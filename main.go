package main

import (
	"context"
	"log"
	"resume-service/internal/auth"
	"resume-service/internal/clients/email"
	"resume-service/internal/clients/filestore"
	"resume-service/internal/clients/mlclient"
	"resume-service/internal/database"
	"resume-service/internal/resume"
	"resume-service/internal/user"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".keys")
	if err != nil {
		log.Fatal("Cannot load env file", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := database.NewClient(ctx)
	if err != nil {
		log.Fatal("Cannot connect to DB", err)
	}
	defer func() { _ = store.Disconnect(ctx) }()

	fileStore := filestore.NewStorageClient()
	mlClient := mlclient.NewMLClient()
	if mlClient == nil {
		log.Fatal("Cannot create ML client")
	}

	mailClient, err := email.NewClient()
	if mailClient == nil {
		log.Fatal("Cannot create mail client", err)
	}

	// Initialize Gin
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Initialize controllers
	userController := user.NewUserController(&store.User, mailClient)
	resumeController := resume.NewResumeController(fileStore, &store.Resume, mlClient)

	// Set up routes
	publicRoutes := r.Group("/api")
	{
		publicRoutes.POST("/signup", userController.Signup)
		publicRoutes.POST("/login", userController.Login)
	}

	authRoutes := r.Group("/api", auth.Middleware())
	{
		authRoutes.POST("/logout", userController.Logout)
		authRoutes.POST("/verify-email", userController.VerifyEmail)
		authRoutes.GET("/resend-otp", userController.ResendOTP)
		authRoutes.PUT("/upload-resume", resumeController.UploadResume)
		authRoutes.GET("/list-resumes", resumeController.ListResumes)
		authRoutes.GET("/download-resume/:resume_id", resumeController.DownloadResume)
		authRoutes.DELETE("/delete-resume/:resume_id", resumeController.DeleteResume)
		authRoutes.POST("/update-resume-visibility/:resume_id", resumeController.UpdateResumeVisibility)
		authRoutes.POST("/generate-cover-letter", resumeController.GenerateCoverletter)
	}

	// Start server
	err = r.Run("localhost:8080")
	if err != nil {
		log.Fatal("Error starting server", err)
	}
}
