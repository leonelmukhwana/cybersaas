package auth

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid registration data",
			"error":   err.Error(),
		})
		return
	}

	result, err := h.service.Register(
		c.Request.Context(),
		req,
	)

	if err != nil {
		if errors.Is(err, ErrEmailExists) {
			c.JSON(http.StatusConflict, gin.H{
				"message": "email already registered",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid login data",
		})
		return
	}

	result, err := h.service.Login(
		c.Request.Context(),
		req,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid email",
		})
		return
	}

	_, err := h.service.ForgotPassword(
		c.Request.Context(),
		req.Email,
	)

	if err != nil {
		log.Printf("Password reset error: %v", err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "unable to process password reset request",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If an account exists with that email, a password reset link has been sent.",
	})
}
func (h *Handler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid reset request",
		})
		return
	}

	if err := h.service.ResetPassword(
		c.Request.Context(),
		req,
	); err != nil {
		if errors.Is(err, ErrInvalidResetToken) {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "reset link is invalid or expired",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "unable to reset password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "password reset successfully",
	})
}
