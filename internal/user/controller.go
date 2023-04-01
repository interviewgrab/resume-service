// TODO: return text error strings instead of errors received from bCrypt / Database
package user

import (
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"resume-service/internal/auth"
	"resume-service/internal/database"
	"resume-service/internal/model"
	"resume-service/internal/utils"
)

type UserController struct {
	userStore *database.UserStore
}

func NewUserController(store *database.UserStore) *UserController {
	return &UserController{userStore: store}
}

func (uc *UserController) Signup(c *gin.Context) {
	var newUser model.User

	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, utils.GinError(err))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	newUser.Password = string(hashedPassword)
	newUser, err = uc.userStore.CreateUser(c, newUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.GinError(err))
		return
	}

	token, err := auth.GenerateJWTToken(newUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	c.JSON(http.StatusCreated, ginToken(token))
}

func (uc *UserController) Login(c *gin.Context) {
	var loginData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, utils.GinError(err))
		return
	}

	user, err := uc.userStore.GetUserByEmail(c, loginData.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.GinError(err))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginData.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, utils.GinError(err))
		return
	}

	token, err := auth.GenerateJWTToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.GinError(err))
		return
	}

	c.JSON(http.StatusOK, ginToken(token))
}

func (uc *UserController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func ginToken(token string) gin.H {
	return gin.H{"token": token}
}
