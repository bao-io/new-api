package controller

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"one-api/common"
	"one-api/i18n"
	"one-api/model"
	"sort"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func TelegramBind(c *gin.Context) {
	if !common.TelegramOAuthEnabled {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage("telegram_oauth_disabled", i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}
	params := c.Request.URL.Query()
	if !checkTelegramAuthorization(params, common.TelegramBotToken) {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage("invalid_request", i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}
	telegramId := params["id"][0]
	if model.IsTelegramIdAlreadyTaken(telegramId) {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage("telegram_id_taken", i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}

	session := sessions.Default(c)
	id := session.Get("id")
	user := model.User{Id: id.(int)}
	if err := user.FillUserById(); err != nil {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage(err.Error(), i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}
	user.TelegramId = telegramId
	if err := user.Update(false); err != nil {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage(err.Error(), i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}

	c.Redirect(302, "/setting")
}

func TelegramLogin(c *gin.Context) {
	if !common.TelegramOAuthEnabled {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage("telegram_oauth_disabled", i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}
	params := c.Request.URL.Query()
	if !checkTelegramAuthorization(params, common.TelegramBotToken) {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage("invalid_request", i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}

	telegramId := params["id"][0]
	user := model.User{TelegramId: telegramId}
	if err := user.FillUserByTelegramId(); err != nil {
		c.JSON(200, gin.H{
			"message": i18n.GetErrorMessage(err.Error(), i18n.GetPreferredLanguage(c)),
			"success": false,
		})
		return
	}
	setupLogin(&user, c)
}

func checkTelegramAuthorization(params map[string][]string, token string) bool {
	strs := []string{}
	var hash = ""
	for k, v := range params {
		if k == "hash" {
			hash = v[0]
			continue
		}
		strs = append(strs, k+"="+v[0])
	}
	sort.Strings(strs)
	var imploded = ""
	for _, s := range strs {
		if imploded != "" {
			imploded += "\n"
		}
		imploded += s
	}
	sha256hash := sha256.New()
	io.WriteString(sha256hash, token)
	hmachash := hmac.New(sha256.New, sha256hash.Sum(nil))
	io.WriteString(hmachash, imploded)
	ss := hex.EncodeToString(hmachash.Sum(nil))
	return hash == ss
}
