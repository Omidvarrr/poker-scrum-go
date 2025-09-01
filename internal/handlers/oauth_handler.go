package handlers

import (
	"awesomeProject1/internal/config"
	"awesomeProject1/internal/models"
	"awesomeProject1/internal/repositories"
	"awesomeProject1/internal/utils"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
)

type OAuthHandler struct {
	cfg      *config.Config
	userRepo repositories.UserRepository
	sessions *session.Store
}

func NewOAuthHandler(cfg *config.Config, userRepo repositories.UserRepository) *OAuthHandler {
	return &OAuthHandler{cfg: cfg, userRepo: userRepo}
}

func (h *OAuthHandler) WithSessionStore(store *session.Store) *OAuthHandler {
	h.sessions = store
	return h
}

func (h *OAuthHandler) GoogleLogin(c *fiber.Ctx) error {
	returnURL := c.Query("return_url", h.cfg.AppBaseURL+"/home")
	// Persist short sid and return_url in server-side session
	sess, err := h.sessions.Get(c)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("session error")
	}
	sid := utils.RandomURLSafe(12)
	sess.Set("sid", sid)
	if len(returnURL) > 512 {
		returnURL = returnURL[:512]
	}
	sess.Set("oauth_return_url", returnURL)
	if err := sess.Save(); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("session save error")
	}
	// Encrypt minimal state (only sid) to keep URL short but confidential
	encState, err := utils.EncryptState(utils.OAuthState{SessionID: sid})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("state error")
	}

	params := url.Values{}
	params.Set("client_id", h.cfg.GoogleClientID)
	params.Set("redirect_uri", h.cfg.GoogleRedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", encState)
	params.Set("access_type", "offline")
	params.Set("prompt", "consent")

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()
	return c.Redirect(authURL, http.StatusFound)
}

func (h *OAuthHandler) GoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" || state == "" {
		return c.Status(fiber.StatusBadRequest).SendString("invalid oauth response")
	}
	// Decrypt state and compare sid with session
	st, err := utils.DecryptState(state)
	if err != nil || st.SessionID == "" {
		return c.Status(fiber.StatusBadRequest).SendString("state invalid")
	}
	sess, err := h.sessions.Get(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("session missing")
	}
	sidVal := sess.Get("sid")
	sidServer, _ := sidVal.(string)
	if sidServer == "" || sidServer != st.SessionID {
		return c.Status(fiber.StatusBadRequest).SendString("state mismatch")
	}
	retVal := sess.Get("oauth_return_url")
	ret, _ := retVal.(string)
	// clear one-time entries
	sess.Delete("sid")
	sess.Delete("oauth_return_url")
	_ = sess.Save()

	// exchange code
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", h.cfg.GoogleClientID)
	form.Set("client_secret", h.cfg.GoogleClientSecret)
	form.Set("redirect_uri", h.cfg.GoogleRedirectURL)
	form.Set("grant_type", "authorization_code")

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("token exchange failed")
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusBadGateway).SendString("token exchange error")
	}
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		IdToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
	}
	_ = json.Unmarshal(body, &tokenResp)

	// fetch user info
	uiReq, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	uiReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	uiResp, err := http.DefaultClient.Do(uiReq)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).SendString("userinfo fetch failed")
	}
	defer uiResp.Body.Close()
	uiBody, _ := io.ReadAll(uiResp.Body)
	if uiResp.StatusCode != http.StatusOK {
		return c.Status(fiber.StatusBadGateway).SendString("userinfo error")
	}
	var userInfo struct {
		Email     string `json:"email"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Picture   string `json:"picture"`
	}
	_ = json.Unmarshal(uiBody, &userInfo)

	// Download and save avatar locally
	localAvatarPath, err := utils.DownloadAndSaveAvatar(userInfo.Picture)
	if err != nil {
		// Log error but don't fail the login process
		localAvatarPath = userInfo.Picture // fallback to original URL
	}

	// find or create user
	user, err := h.userRepo.GetUserByEmail(userInfo.Email)
	if err != nil {
		user, err = h.userRepo.CreateUser(models.User{Email: userInfo.Email, FirstName: userInfo.FirstName, LastName: userInfo.LastName, Avatar: localAvatarPath, ProfileCompleted: true})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString("user create failed")
		}
	} else {
		_ = h.userRepo.CompleteUserProfile(user.ID, userInfo.FirstName, userInfo.LastName, localAvatarPath)
	}

	// issue our tokens in cookies
	access, err := utils.GenerateToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("token fail")
	}
	refresh, err := utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("token fail")
	}

	c.Cookie(&fiber.Cookie{Name: "token", Value: access, HTTPOnly: true, Secure: false, SameSite: "Lax", Expires: time.Now().Add(24 * time.Hour)})
	c.Cookie(&fiber.Cookie{Name: "refresh_token", Value: refresh, HTTPOnly: true, Secure: false, SameSite: "Lax", Expires: time.Now().Add(7 * 24 * time.Hour)})

	if ret == "" {
		ret = h.cfg.AppBaseURL + "/home"
	}
	return c.Redirect(ret, http.StatusFound)
}

func (h *OAuthHandler) Logout(c *fiber.Ctx) error {
	// Clear the authentication cookies
	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(-time.Hour), // Set to past time to delete
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
		Expires:  time.Now().Add(-time.Hour), // Set to past time to delete
	})

	return c.JSON(fiber.Map{"success": true})
}
