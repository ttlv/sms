package action

import (
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/common_utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/internal"
	"github.com/ttlv/sms/queue/amqp_queue"
	"github.com/ttlv/sms/server"
	"github.com/ttlv/sms/service/producer_http/app/helpers"
	"net/http"
)

type Handlers struct {
	DB           *gorm.DB
	SessionStore *sessions.CookieStore
	AMQPConn     *amqp.Connection
	Channel      *amqp.Channel
}

func NewHandlers(db *gorm.DB, sessionStore *sessions.CookieStore, AMQPConn *amqp.Connection, channel *amqp.Channel) Handlers {
	return Handlers{DB: db, SessionStore: sessionStore, AMQPConn: AMQPConn, Channel: channel}
}

func (handler *Handlers) Send(w http.ResponseWriter, r *http.Request) {
	var (
		smsServer internal.SmsServer
		err       error
		params    sms.SendParams
		brand     = r.PostFormValue("brand")
		phone     = r.PostFormValue("phone")
		content   = r.PostFormValue("content")
		b         = sms.SmsBrand{}
		amount    = common_utils.Count{}
	)
	// 调用api之前再进行一次权限校验
	if _, err := helpers.GetToken(r, handler.SessionStore); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
	if smsServer, err = server.New(handler.DB, amqp_queue.New(handler.AMQPConn, handler.Channel), []sms.SmsProvider{}); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
	if brand == "" {
		helpers.RenderFailureJSON(w, 400, "未填写Brand")
		return
	}
	handler.DB.First(&b, "name = ?", brand)
	if handler.DB.NewRecord(&b) {
		helpers.RenderFailureJSON(w, 400, "无效brand")
		return
	}
	// 判断该brand是否还有可用的短信条数
	handler.DB.Raw(`SELECT SUM(available_amount) AS value FROM sms_availables WHERE sms_brand_id = ?`, b.ID).Scan(&amount)
	if amount.Value <= 0 {
		helpers.RenderFailureJSON(w, 400, "您无可用的短信条数,请联系管理员充值.")
		return
	}
	if phone == "" {
		helpers.RenderFailureJSON(w, 400, "未填写手机号码")
		return
	}
	if content == "" {
		helpers.RenderFailureJSON(w, 400, "未填写需要发送的内容")
		return
	}
	params.Brand = brand
	params.Phone = phone
	params.Content = content
	if err = smsServer.HttpSend(&params); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
}
