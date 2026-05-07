package handlers

import (
	stdErrors "errors"
	"log"
	"net/http"

	"github.com/amirrajj-dev/taskio/internal/configs"
	"github.com/amirrajj-dev/taskio/internal/dtos"
	appErr "github.com/amirrajj-dev/taskio/internal/errors"
	"github.com/amirrajj-dev/taskio/internal/events"
	"github.com/amirrajj-dev/taskio/internal/helpers"
	"github.com/amirrajj-dev/taskio/internal/infrastructure/queue"
	"github.com/amirrajj-dev/taskio/internal/models"
	"github.com/amirrajj-dev/taskio/internal/services"
	"github.com/amirrajj-dev/taskio/internal/utils"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService services.AuthService
	activityProducer *queue.ActivityProducer
}

func NewAuthHandler(producer *queue.ActivityProducer) *AuthHandler {
	return &AuthHandler{
		authService: services.UserService,
		activityProducer: producer,
	}
}

// Register godoc
// @Summary      User registration
// @Description  Registers a new user account
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.RegisterRequest  true  "User registration details"
// @Success      201  {object}  utils.SuccessResponse{data=models.UserResponse}  "Registered successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dtos.RegisterRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	createdUser, token, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "user already exists" {
			c.JSON(http.StatusConflict, appErr.NewBasicError("user already exists", c.Request.URL.Path))
			return
		}
		c.JSON(http.StatusInternalServerError, appErr.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	var isSecure bool
	if configs.Configs.App.GO_ENV == "production" {
		isSecure = true
	} else {
		isSecure = false
	}
	cookie := &http.Cookie{
		Name:     configs.Configs.COOKIE_NAME,
		Path:     "/",
		Value:    *token,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(configs.Configs.JWT.JWT_EXPIRY_HOURS_REGISTER.Seconds()),
	}
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	c.JSON(http.StatusCreated, utils.NewSuccessResponse("registered succesfully", createdUser, c.Request.URL.Path))
}

// Login godoc
// @Summary      User login
// @Description  Authenticates a user and returns a JWT token in cookie
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      dtos.LoginRequest  true  "Login credentials"
// @Success      200  {object}  utils.SuccessResponse{data=models.UserResponse}  "Logged in successfully"
// @Failure      400  {object}  errors.BasicError  "Validation error"
// @Failure      401  {object}  errors.BasicError  "Invalid credentials"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dtos.LoginRequest
	if !helpers.ShouldBindJSON(c, &req) {
		return
	}
	if validationErr := utils.ValidateRequest(c, req); validationErr != nil {
		return
	}
	user, token, err := h.authService.Login(c, req)
	if err != nil {
		if err.Error() == "invalid credentials" || stdErrors.Is(err, appErr.ErrUserNotFound) {
			c.JSON(http.StatusUnauthorized, appErr.NewBasicError("invalid credentials", c.Request.URL.Path))
			return
		}
		c.JSON(http.StatusInternalServerError, appErr.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	var isSecure bool
	if configs.Configs.App.GO_ENV == "production" {
		isSecure = true
	} else {
		isSecure = false
	}
	cookie := &http.Cookie{
		Name:     configs.Configs.COOKIE_NAME,
		Path:     "/",
		Value:    *token,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge: int(configs.Configs.JWT.JWT_EXPIRY_HOURS_LOGIN.Seconds()),
	}
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	c.JSON(http.StatusOK, utils.NewSuccessResponse("logged in succesfully", user, c.Request.URL.Path))
}

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Refreshes the authentication token using the refresh token in cookie
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponse  "Token refreshed successfully"
// @Failure      401  {object}  errors.BasicError  "Unauthorized"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, appErr.NewBasicError("unautorized", c.Request.URL.Path))
		return
	}
	refreshToken, err := h.authService.RefreshToken(c.Request.Context(), *userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, appErr.NewBasicError(err.Error(), c.Request.URL.Path))
		return
	}
	var isSecure bool
	if configs.Configs.App.GO_ENV == "production" {
		isSecure = true
	} else {
		isSecure = false
	}
	cookie := &http.Cookie{
		Name:     configs.Configs.COOKIE_NAME,
		Path:     "/",
		Value:    refreshToken.Token,
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(configs.Configs.JWT.JWT_EXPIRY_HOURS_REFRESH.Seconds()),
	}
	c.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
	c.JSON(http.StatusOK, utils.NewSuccessResponse("token refreshed succesfully", nil, c.Request.URL.Path))
}

// LogOut godoc
// @Summary      User logout
// @Description  Invalidates the refresh token and clears the authentication cookie
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  utils.SuccessResponse  "Logged out successfully"
// @Failure      500  {object}  errors.BasicError  "Internal server error"
// @Router       /auth/logout [post]
func (h *AuthHandler) LogOut(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		utils.ClearAuthCookie(c)
		c.JSON(http.StatusOK, utils.NewSuccessResponse("logged out succesfully", nil, c.Request.URL.Path))
		return
	}
	refreshToken, findErr := h.authService.FindRefreshTokenByUserID(c.Request.Context(), *userID)
	if findErr != nil {
		if findErr.Error() != "refresh not found" {
			log.Printf("Error finding refresh token for user %s during logout: %v", userID.String(), findErr)
		}
	} else {
		if deleteErr := h.authService.DeleteRefresh(c.Request.Context(), refreshToken.ID); deleteErr != nil {
			if deleteErr.Error() == "refresh not found" {
				log.Printf("refresh not found for user %s during logout", userID.String())
			} else {
				log.Printf("Error deleting refresh token %s for user %s during logout: %v", refreshToken.ID.String(), userID.String(), deleteErr)
			}
		}
	}
	utils.ClearAuthCookie(c)
	event := models.ActivityEvent{
		EventName: events.ActivityEvents.LogOut,
		UserID: userID,
	}
	if err := h.activityProducer.Publish(event);err != nil {
		log.Printf("failed to publish logout event => %v" , err)
	}
	c.JSON(http.StatusOK, utils.NewSuccessResponse("logged out succesfully", nil, c.Request.URL.Path))

}
