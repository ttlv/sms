package webhook

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	"github.com/tidwall/gjson"
	"github.com/ttacon/libphonenumber"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
	"github.com/ttlv/sms/queue/amqp_queue"
)

type YunPianResp struct {
	Sid             int64  `json:"sid"`
	Uid             string `json:"uid"`
	UserReceiveTime string `json:"user_receive_time"`
	ErrorMsg        string `json:"error_msg"`
	Mobile          string `json:"mobile"`
	ReportStatus    string `json:"report_status"`
}

type statusReport struct {
	Mobile       string `json:"mobile"`
	SmsID        string `json:"smsId"`
	CustomSmsID  string `json:"customSmsId"`
	State        string `json:"state"`
	Desc         string `json:"desc"`
	ReceiveTime  string `json:"receiveTime"`
	SubmitTime   string `json:"submitTime"`
	ExtendedCode string `json:"extendedCode"`
}

type WebHookServer struct {
	DB        *gorm.DB
	AmqpQueue *amqp_queue.AMQPQueue
}

func NewWebHookServer(db *gorm.DB, amqpQueue *amqp_queue.AMQPQueue) WebHookServer {
	return WebHookServer{DB: db, AmqpQueue: amqpQueue}
}

func (whs WebHookServer) GetYunPianCallBack(w http.ResponseWriter, req *http.Request) {
	var resp YunPianResp
	raw, _ := ioutil.ReadAll(req.Body)
	Log("provider=yunpian body=%s", string(raw))
	s, _ := url.ParseQuery(string(raw))
	urlDecode, _ := url.QueryUnescape(s.Get("sms_status"))
	json.Unmarshal([]byte(urlDecode), &resp)
	record := sms.SmsRecord{}
	whs.DB.Where("external_id = ?", resp.Sid).First(&record)

	t := time.Now()
	record.LastCallbackAt = &t
	if !whs.DB.NewRecord(record) {
		if resp.ReportStatus == "SUCCESS" {
			record.State = sms.RecordState_Delivered
			record.Error = ""
			whs.DB.Save(&record)
		} else {
			whs.HandleError(&record, "YunPian", resp.ErrorMsg)
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS"))
}

func (whs WebHookServer) GetTwilioCallBack(w http.ResponseWriter, req *http.Request) {
	record := sms.SmsRecord{}
	raw, _ := ioutil.ReadAll(req.Body)
	Log("provider=twilio body=%s", string(raw))
	s, _ := url.ParseQuery(string(raw))
	smsSid, _ := url.QueryUnescape(s.Get("SmsSid"))
	messageStatus, _ := url.QueryUnescape(s.Get("SmsStatus"))
	whs.DB.Where("external_id = ?", smsSid).First(&record)
	t := time.Now()
	record.LastCallbackAt = &t
	if !whs.DB.NewRecord(record) {
		if messageStatus == "delivered" {
			record.State = sms.RecordState_Delivered
			record.Error = ""
			whs.DB.Save(&record)
		} else if messageStatus == "failed" || messageStatus == "undelivered" {
			errorMessage, _ := url.QueryUnescape(s.Get("ErrorMessage"))
			whs.HandleError(&record, "Twilio", errorMessage)
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("SUCCESS"))
}

func (whs WebHookServer) GetEmayCallBack(w http.ResponseWriter, req *http.Request) {
	var (
		errMsg string
	)
	reports := make([]statusReport, 0)
	raw, _ := ioutil.ReadAll(req.Body)
	Log("provider=emay body=%v", string(raw))
	jsonStr := strings.Replace(string(raw), "reports=", "", 1)
	json.Unmarshal([]byte(jsonStr), &reports)
	for _, report := range reports {
		record := sms.SmsRecord{}
		whs.DB.Where("external_id = ?", report.SmsID).First(&record)
		t := time.Now()
		record.LastCallbackAt = &t
		if !whs.DB.NewRecord(&record) {
			if report.State == "DELIVRD" {
				record.State = sms.RecordState_Delivered
				record.Error = ""
				whs.DB.Save(&record)
			} else {
				record.State = sms.RecordState_Failure
				if _, ok := EmayErrorMsg[report.Desc]; ok {
					errMsg = EmayErrorMsg[report.Desc]
				} else {
					errMsg = report.Desc
				}
				whs.HandleError(&record, "Emay", errMsg)
			}
		}
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

func (whs WebHookServer) GetAmazonCallBack(w http.ResponseWriter, req *http.Request) {
	raw, _ := ioutil.ReadAll(req.Body)
	body := string(raw)
	Log("provider=amazon body=%v", body)
	record := sms.SmsRecord{}
	message := gjson.Get(body, "logEvents.0.message").String()
	whs.DB.Where("external_id = ?", gjson.Get(message, "notification.messageId").String()).First(&record)
	t := time.Now()
	record.LastCallbackAt = &t
	if !whs.DB.NewRecord(&record) {
		status := gjson.Get(message, "status").String()
		if status == "SUCCESS" {
			record.State = sms.RecordState_Delivered
			record.Error = ""
			whs.DB.Save(&record)
		} else {
			whs.HandleError(&record, "Amazon", gjson.Get(message, "delivery.providerResponse").String())
		}
	}
}

func (whs WebHookServer) GetChuangLanCallBack(w http.ResponseWriter, req *http.Request) {
	var (
		values            url.Values
		chuangLanCallBack = sms.ChuangLanCallBack{}
		record            = sms.SmsRecord{}
	)
	values, _ = url.ParseQuery(strings.Replace(req.URL.String(), "/chuanglan?", "", -1))
	whs.DB.Where("external_id = ?", values.Get("msgid")).First(&record)
	if values.Get("status") == "DELIVRD" {
		record.State = sms.RecordState_Delivered
		record.Error = ""
		whs.DB.Save(&record)
	} else {
		whs.HandleError(&record, "ChuangLan", values.Get("statusDesc"))
	}
	chuangLanCallBack.Clcode = "000000"
	jsonData, _ := json.Marshal(chuangLanCallBack)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
func (whs WebHookServer) Liveness(w http.ResponseWriter, req *http.Request) {
	healthCheck := struct {
		Healthy bool   `json:"healthy"`
		Code    string `json:"code"`
	}{
		Healthy: true, Code: "200",
	}
	jsonData, _ := json.Marshal(healthCheck)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

}

func (whs WebHookServer) Run() {
	router := mux.NewRouter()
	c := config.MustGetConfig()
	router.HandleFunc("/yunpian", whs.GetYunPianCallBack).Methods("POST")
	router.HandleFunc("/twilio", whs.GetTwilioCallBack).Methods("POST")
	router.HandleFunc("/emay", whs.GetEmayCallBack).Methods("POST")
	router.HandleFunc("/aws", whs.GetAmazonCallBack).Methods("POST")
	router.HandleFunc("/chuanglan", whs.GetChuangLanCallBack).Methods("GET")
	router.HandleFunc("/liveness", whs.Liveness).Methods("Get")
	http.ListenAndServe(c.Port, router)
}

func Log(format string, data ...interface{}) {
	timezone := time.FixedZone("local", 8*60*60)
	msg := fmt.Sprintf(format, data...)
	fmt.Printf("service=smswebhook time=%v %v\n", time.Now().In(timezone).Unix(), msg)
}
func (whs WebHookServer) HandleError(smsRecord *sms.SmsRecord, providerName string, errMsg string) {
	var (
		publishData       = &sms.PublishData{}
		sendParams        = sms.SendParams{}
		smsFailureRecords = []sms.SmsFailureRecord{}
	)
	json.Unmarshal([]byte(smsRecord.RawParam), &sendParams)

	// 保存错误
	smsRecord.State = sms.RecordState_Failure
	smsRecord.Error = smsRecord.Error + fmt.Sprintf("%v: %v; ", providerName, errMsg)
	whs.DB.Save(smsRecord)

	// 创建失败记录
	rawPhone := sendParams.Phone
	if phoneNumber, err := libphonenumber.Parse(sendParams.Phone, libphonenumber.UNKNOWN_REGION); err == nil {
		rawPhone = libphonenumber.Format(phoneNumber, libphonenumber.INTERNATIONAL)
	}
	whs.DB.Save(&sms.SmsFailureRecord{
		SmsRecordId:  smsRecord.ID,
		ProviderName: providerName,
		Phone:        rawPhone,
		Error:        errMsg,
	})

	// 重新发送
	publishData.SmsRecordId = smsRecord.ID
	whs.DB.First(&smsFailureRecords, "sms_record_id = ?", smsRecord.ID)
	for _, s := range smsFailureRecords {
		publishData.SentProviders = append(publishData.SentProviders, s.ProviderName)
	}
	publishData.SentProviders = append(publishData.SentProviders, providerName)
	publishData.SendParams = &sms.SendParams{
		Brand:   sendParams.Brand,
		Phone:   sendParams.Phone,
		Country: sendParams.Country,
		Content: sendParams.Content,
	}
	whs.AmqpQueue.Publish(publishData)
}
