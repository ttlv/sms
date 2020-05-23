package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/service/producer_http/app/entries"
	"net/http"
)

func RenderFailureJSON(w http.ResponseWriter, code int, message string) {
	result, _ := json.Marshal(entries.Error{
		Error: entries.ErrorDetail{
			Message: message,
			Code:    code,
		},
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func RenderSuccessJSON(w http.ResponseWriter, data interface{}) {
	result, _ := json.Marshal(entries.Success{
		Data: data,
	})
	w.Header().Set("Content-Type", "application/json")
	w.Write(result)
}

func GetToken(db *gorm.DB, brandID uint, r *http.Request, sessionStore *sessions.CookieStore) (string, error) {
	var (
		smsBrand    = sms.SmsBrand{}
		accessToken string
		ok          bool
	)
	session, err := sessionStore.Get(r, "sms_service")
	if err != nil {
		return "", err
	}
	if accessToken, ok = session.Values["token"].(string); ok {
		// 从db中对比token是否正确
		db.First(&smsBrand, "id = ?", brandID)
		if !db.NewRecord(&smsBrand) && smsBrand.Token == accessToken {
			return accessToken, nil
		}
	}
	return "", fmt.Errorf("未发现token,无法调用服务。")
}
