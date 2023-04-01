package main

import (
	"context"
	"log"
	"resume-service/internal/auth"
	"resume-service/internal/clients/filestore"
	"resume-service/internal/database"
	"resume-service/internal/resume"
	"resume-service/internal/user"
	"time"

	"github.com/gin-gonic/gin"
)

const ()

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	store, err := database.NewClient(ctx)
	if err != nil {
		log.Fatal("Cannot connect to DB", err)
	}
	defer func() { _ = store.Disconnect(ctx) }()

	fileStore := filestore.NewStorageClient()

	// Initialize Gin
	r := gin.Default()

	// Initialize controllers
	userController := user.NewUserController(&store.User)
	resumeController := resume.NewResumeController(fileStore, &store.Resume)

	// Set up routes
	publicRoutes := r.Group("/api")
	{
		publicRoutes.POST("/signup", userController.Signup)
		publicRoutes.POST("/login", userController.Login)
	}

	authRoutes := r.Group("/api", auth.Middleware())
	{
		authRoutes.PUT("/upload-resume", resumeController.UploadResume)
		authRoutes.GET("/list-resumes", resumeController.ListResumes)
		authRoutes.GET("/download-resume/:resume_id", resumeController.DownloadResume)
		authRoutes.POST("/mark-resume-public/:resume_id", resumeController.MarkResumePublic)
		authRoutes.POST("/logout", userController.Logout)
	}

	// Start server
	err = r.Run("localhost:8080")
	if err != nil {
		log.Fatal("Error starting server", err)
	}
}
