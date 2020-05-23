package action

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
	"github.com/ttlv/common_utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/internal"
	"github.com/ttlv/sms/queue/amqp_queue"
	"github.com/ttlv/sms/server"
	"github.com/ttlv/sms/service/producer_http/app/helpers"
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
		b         = sms.SmsBrand{}
		amount    = common_utils.Count{}
		apiParams = sms.ApiParams{}
	)
	if smsServer, err = server.New(handler.DB, amqp_queue.New(handler.AMQPConn, handler.Channel), []sms.SmsProvider{}); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
	requestBody, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(requestBody, &apiParams)
	if apiParams.Brand == "" {
		helpers.RenderFailureJSON(w, 400, "未填写Brand")
		return
	}
	handler.DB.First(&b, "name = ?", apiParams.Brand)
	if handler.DB.NewRecord(&b) {
		helpers.RenderFailureJSON(w, 400, "无效brand")
		return
	}
	// 调用api之前进行一次token校验
	if _, err := helpers.GetToken(handler.DB, b.ID, r, handler.SessionStore); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
	if apiParams.Phone == "" {
		helpers.RenderFailureJSON(w, 400, "未填写手机号码")
		return
	}
	if apiParams.Content == "" {
		helpers.RenderFailureJSON(w, 400, "未填写需要发送的内容")
		return
	}
	// 判断该brand是否还有可用的短信条数
	handler.DB.Raw(`SELECT SUM(available_amount) AS value FROM sms_availables WHERE sms_brand_id = ?`, b.ID).Scan(&amount)
	if amount.Value <= 0 {
		helpers.RenderFailureJSON(w, 400, "您无可用的短信条数,请联系管理员充值.")
		return
	}
	params.Brand = apiParams.Brand
	params.Phone = apiParams.Phone
	params.Content = apiParams.Content
	if err = smsServer.HttpSend(&params); err != nil {
		helpers.RenderFailureJSON(w, 400, err.Error())
		return
	}
}
