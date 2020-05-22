package helpers

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
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

func GetToken(r *http.Request, sessionStore *sessions.CookieStore) (string, error) {
	session, err := sessionStore.Get(r, "sms_service")
	if err != nil {
		return "", err
	}
	if accessToken, ok := session.Values["token"].(string); ok {
		return accessToken, nil
	}
	return "", fmt.Errorf("未发现token,无法调用服务。")
}
