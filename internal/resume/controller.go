package resume

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"io"
	"net/http"
	"resume-service/internal/auth"
	"resume-service/internal/clients/filestore"
	"resume-service/internal/clients/mlclient"
	"resume-service/internal/database"
	"resume-service/internal/model"
	"resume-service/internal/utils"
	"strconv"
	"time"
)

type ResumeController struct {
	fileStorage *filestore.FileStore
	resumeStore *database.ResumeStore
	mlclient    *mlclient.MLClient
}

func NewResumeController(fileStorage *filestore.FileStore, store *database.ResumeStore, mlclient *mlclient.MLClient) *ResumeController {
	return &ResumeController{fileStorage: fileStorage, resumeStore: store, mlclient: mlclient}
}

func (r *ResumeController) UploadResume(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	key := fmt.Sprintf("user-%s-%s", auth.GetUserIdFromContext(c).String(), uuid.New())

	err = r.fileStorage.Upload(key, fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}
	resume := model.Resume{
		UserID:     auth.GetUserIdFromContext(c),
		FileName:   fileHeader.Filename,
		Key:        key,
		UploadDate: time.Now(),
		Metadata:   map[string]string{},
		Public:     false,
	}

	err = r.resumeStore.StoreResume(c, resume)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"result": "Resume uploaded successfully"})
}

func (r *ResumeController) ListResumes(c *gin.Context) {
	userId := auth.GetUserIdFromContext(c)

	resumes, err := r.resumeStore.GetResumesByUserId(c, userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"resumes": resumes})
}

func (r *ResumeController) DownloadResume(c *gin.Context) {
	resumeId := c.Param("resume_id")
	if resumeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Now resume id found"})
		return
	}

	resume, err := r.resumeStore.GetResume(c, resumeId)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.GinError(err))
		return
	}

	userId := auth.GetUserIdFromContext(c)
	if !resume.Public && resume.UserID != userId {
		c.JSON(http.StatusUnauthorized, utils.GinError(errors.New("not allowed to download resume")))
		return
	}

	file, err := r.fileStorage.Download(resume.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+resume.FileName)
	c.Data(http.StatusOK, "application/pdf", file)
}

func (r *ResumeController) UpdateResumeVisibility(c *gin.Context) {
	userId := auth.GetUserIdFromContext(c)
	resumeId := c.Param("resume_id")
	if resumeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resume_id not found"})
		return
	}

	isPublic, err := strconv.ParseBool(c.DefaultQuery("public", "false"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	err = r.resumeStore.UpdateUserResumeIsPublic(c, userId, resumeId, isPublic)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Resume visibility updated"})
}

func (r *ResumeController) GenerateCoverletter(c *gin.Context) {
	var request struct {
		ResumeId string `json:"resume_id"`
		JobDesc  string `json:"job_desc"`
	}

	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, utils.GinError(err))
		return
	}

	if request.ResumeId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume ID and Job Description are required"})
		return
	}

	resume, err := r.resumeStore.GetResume(c, request.ResumeId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	// verify if user can read this resume
	if resume.UserID != auth.GetUserIdFromContext(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// download PDF from S3
	fileContent, err := r.fileStorage.Download(resume.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	// Extract text from the PDF
	resumeText, err := parsePDF(fileContent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse resume text"})
		return
	}

	// Generate the cover letter
	coverLetter, err := r.mlclient.GenerateCoverLetter(c, request.JobDesc, resumeText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate cover letter"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cover_letter": coverLetter})
}
