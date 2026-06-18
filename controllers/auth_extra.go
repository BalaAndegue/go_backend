package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
	"shopcart-api/utils"
)

// exposeTokens reports whether reset/verification tokens should be returned in
// API responses. Useful when no email provider is wired up. Defaults to true.
func exposeTokens() bool {
	v := getenvDefault("EXPOSE_TOKENS", "true")
	return v != "false" && v != "0"
}

// issueVerificationToken creates a single-use token for the given purpose,
// invalidating any prior unused tokens of the same purpose, and returns the
// raw token (only its hash is stored).
func issueVerificationToken(userID uint, purpose string, ttl time.Duration) (string, error) {
	raw, err := utils.GenerateRandomToken()
	if err != nil {
		return "", err
	}
	// Drop previous outstanding tokens of the same purpose for this user.
	config.DB.Where("user_id = ? AND purpose = ? AND used_at IS NULL", userID, purpose).
		Delete(&models.VerificationToken{})

	vt := models.VerificationToken{
		UserID:    userID,
		Purpose:   purpose,
		TokenHash: utils.HashToken(raw),
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := config.DB.Create(&vt).Error; err != nil {
		return "", err
	}
	return raw, nil
}

// issueEmailVerifyToken is a convenience wrapper used at registration time.
func issueEmailVerifyToken(userID uint) (string, error) {
	return issueVerificationToken(userID, models.PurposeEmailVerify, 24*time.Hour)
}

// consumeToken validates and marks a token used, returning the owning user.
func consumeToken(raw, purpose string) (*models.User, bool) {
	var vt models.VerificationToken
	err := config.DB.Where(
		"token_hash = ? AND purpose = ? AND used_at IS NULL AND expires_at > ?",
		utils.HashToken(raw), purpose, time.Now(),
	).First(&vt).Error
	if err != nil {
		return nil, false
	}

	var user models.User
	if err := config.DB.First(&user, vt.UserID).Error; err != nil {
		return nil, false
	}
	now := time.Now()
	config.DB.Model(&vt).Update("used_at", now)
	return &user, true
}

// ForgotPassword issues a password-reset token for the given email. The
// response is the same whether or not the email exists, to avoid user
// enumeration (the token is only included when found and EXPOSE_TOKENS is on).
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{"message": "If the email exists, a reset token has been issued."}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err == nil {
		if raw, err := issueVerificationToken(user.ID, models.PurposePasswordReset, time.Hour); err == nil && exposeTokens() {
			resp["reset_token"] = raw
		}
	}
	c.JSON(http.StatusOK, resp)
}

// ResetPassword sets a new password given a valid reset token and invalidates
// all existing sessions by bumping the user's token version.
func ResetPassword(c *gin.Context) {
	var input struct {
		Token                string `json:"token" binding:"required"`
		Password             string `json:"password" binding:"required,min=8"`
		PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	user, ok := consumeToken(input.Token, models.PurposePasswordReset)
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid or expired token"})
		return
	}

	hashed, err := utils.HashPassword(input.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}
	config.DB.Model(user).Updates(map[string]interface{}{
		"password":      hashed,
		"token_version": user.TokenVersion + 1,
	})
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}

// VerifyEmail marks the user's email as verified given a valid token.
func VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var input struct {
			Token string `json:"token"`
		}
		_ = c.ShouldBindJSON(&input)
		token = input.Token
	}
	if token == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "token is required"})
		return
	}

	user, ok := consumeToken(token, models.PurposeEmailVerify)
	if !ok {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Invalid or expired token"})
		return
	}
	now := time.Now()
	config.DB.Model(user).Update("email_verified_at", now)
	c.JSON(http.StatusOK, gin.H{"message": "Email verified successfully"})
}

// ResendVerification issues a fresh email-verification token for the caller.
func ResendVerification(c *gin.Context) {
	user, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthenticated"})
		return
	}
	if user.EmailVerifiedAt != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Email already verified"})
		return
	}
	raw, err := issueVerificationToken(user.ID, models.PurposeEmailVerify, 24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not issue token"})
		return
	}
	resp := gin.H{"message": "Verification token issued"}
	if exposeTokens() {
		resp["verification_token"] = raw
	}
	c.JSON(http.StatusOK, resp)
}
