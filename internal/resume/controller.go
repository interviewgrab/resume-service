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
	"resume-service/internal/database"
	"resume-service/internal/model"
	"resume-service/internal/utils"
	"strconv"
	"time"
)

type ResumeController struct {
	fileStorage *filestore.FileStore
	resumeStore *database.ResumeStore
}

func NewResumeController(fileStorage *filestore.FileStore, store *database.ResumeStore) *ResumeController {
	return &ResumeController{fileStorage: fileStorage, resumeStore: store}
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
