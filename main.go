package main

import (
	"context"
	"log"
	"os"
	"resume-service/internal/auth"
	"resume-service/internal/clients/email"
	"resume-service/internal/clients/filestore"
	"resume-service/internal/clients/mlclient"
	"resume-service/internal/clients/parameters"
	"resume-service/internal/database"
	"resume-service/internal/resume"
	"resume-service/internal/user"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
	KEY_MONGO_URI      = "MONGO_URI"
	KEY_OPENAI_API_KEY = "OPENAI_API_KEY"
	KEY_SENDER_EMAIL   = "SENDER_EMAIL"
	KEY_SENDER_PASS    = "SENDER_PASS"
)

var service_params = []string{KEY_MONGO_URI, KEY_OPENAI_API_KEY, KEY_SENDER_EMAIL, KEY_SENDER_PASS}

func main() {
	err := godotenv.Load(".keys")
	if err != nil {
		log.Fatal("Cannot load env file", err)
	}

	paramClient, err := parameters.NewParamClient(os.Getenv("REGION"))
	if err != nil {
		log.Fatal("Cannot create param client", err)
	}

	for _, param := range service_params {
		paramValue, err := paramClient.GetStringParam(param)
		if err != nil {
			log.Fatal("Cannot read param: %s", err)
		}
		err = os.Setenv(param, paramValue)
		if err != nil {
			log.Fatal("Cannot set param: %s", err)
		}
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
	userPublicRoutes := r.Group("/api")
	{
		userPublicRoutes.POST("/signup", userController.Signup)
		userPublicRoutes.POST("/login", userController.Login)
	}

	userAuthedRoutes := r.Group("/api", auth.Middleware())
	{
		userAuthedRoutes.POST("/logout", userController.Logout)
		userAuthedRoutes.POST("/verify-email", userController.VerifyEmail)
		userAuthedRoutes.GET("/resend-otp", userController.ResendOTP)
	}

	resumeRoutes := r.Group("/api", auth.Middleware(), auth.EmailVerified(&store.User))
	{
		resumeRoutes.PUT("/upload-resume", resumeController.UploadResume)
		resumeRoutes.GET("/list-resumes", resumeController.ListResumes)
		resumeRoutes.GET("/download-resume/:resume_id", resumeController.DownloadResume)
		resumeRoutes.DELETE("/delete-resume/:resume_id", resumeController.DeleteResume)
		resumeRoutes.POST("/update-resume-visibility/:resume_id", resumeController.UpdateResumeVisibility)
		resumeRoutes.POST("/generate-cover-letter", resumeController.GenerateCoverletter)
	}

	// Start server
	err = r.Run("localhost:8080")
	if err != nil {
		log.Fatal("Error starting server", err)
	}
}
