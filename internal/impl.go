package internal

import (
	"context"

	"encoding/json"
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/ttacon/libphonenumber"
	"github.com/ttlv/common_utils/health_checker"
	"github.com/ttlv/sms"
	"time"
)

type SmsServer struct {
	Providers []sms.SmsProvider
	Queue     sms.SmsQueue
	DB        *gorm.DB
}

func New(db *gorm.DB, queue sms.SmsQueue, providers []sms.SmsProvider) (serv SmsServer, err error) {
	server := SmsServer{Providers: providers, Queue: queue, DB: db}
	db.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsBrand{}, &sms.SmsSetting{})
	return server, nil
}
func (ser SmsServer) HttpSend(params *sms.SendParams) (err error) {
	var (
		number string
	)
	_, number, err = ser.parsePhoneNumber(params)
	rawParamByte, _ := json.Marshal(params)
	waitCallbackRecord := sms.SmsRecord{}
	if err == nil {
		smsRecord := sms.SmsRecord{
			Phone:    number,
			Brand:    params.Brand,
			RawParam: string(rawParamByte),
			State:    sms.RecordState_Sending,
		}
		ser.DB.Save(&smsRecord)
		// 创建数据之后判断查找DB中是否存在着上一条是相同用户发送的短信数据，并且已经发送的短信在等到短信服务商的callback,如果存在,这次发送跳过上一条发送的运营商发送
		ser.DB.Where("phone = ? AND state = ?", smsRecord.Phone, sms.RecordState_Success).Last(&waitCallbackRecord)
		if !ser.DB.NewRecord(&waitCallbackRecord) {
			ser.Queue.Publish(&sms.PublishData{SmsRecordId: smsRecord.ID, SendParams: params, SentProviders: []string{waitCallbackRecord.Sender}})
		} else {
			ser.Queue.Publish(&sms.PublishData{SmsRecordId: smsRecord.ID, SendParams: params})
		}
	} else {
		smsRecordData := sms.SmsRecord{
			Phone:    number,
			Brand:    params.Brand,
			RawParam: string(rawParamByte),
			State:    sms.RecordState_Failure,
			Error:    err.Error(),
		}
		ser.DB.Save(&smsRecordData)
	}
	return

}
func (ser SmsServer) Send(context context.Context, params *sms.SendParams) (resp *sms.SendResp, err error) {
	_, number, err := ser.parsePhoneNumber(params)
	rawParamByte, _ := json.Marshal(params)
	waitCallbackRecord := sms.SmsRecord{}
	if err == nil {
		smsRecord := sms.SmsRecord{
			Phone:    number,
			Brand:    params.Brand,
			RawParam: string(rawParamByte),
			State:    sms.RecordState_Sending,
		}
		ser.DB.Save(&smsRecord)
		// 创建数据之后判断查找DB中是否存在着上一条是相同用户发送的短信数据，并且已经发送的短信在等到短信服务商的callback,如果存在,这次发送跳过上一条发送的运营商发送
		ser.DB.Where("phone = ? AND state = ?", smsRecord.Phone, sms.RecordState_Success).Last(&waitCallbackRecord)
		if !ser.DB.NewRecord(&waitCallbackRecord) {
			ser.Queue.Publish(&sms.PublishData{SmsRecordId: smsRecord.ID, SendParams: params, SentProviders: []string{waitCallbackRecord.Sender}})
		} else {
			ser.Queue.Publish(&sms.PublishData{SmsRecordId: smsRecord.ID, SendParams: params})
		}
		return &sms.SendResp{Uid: fmt.Sprintf("%v", smsRecord.ID)}, nil
	} else {
		smsRecordData := sms.SmsRecord{
			Phone:    number,
			Brand:    params.Brand,
			RawParam: string(rawParamByte),
			State:    sms.RecordState_Failure,
			Error:    err.Error(),
		}
		ser.DB.Save(&smsRecordData)
		return &sms.SendResp{Uid: fmt.Sprintf("%v", smsRecordData.ID), Error: smsRecordData.Error}, nil
	}
	return
}

func (ser SmsServer) RealSend(publishData *sms.PublishData) (err error) {
	var providerResp, externalID string
	country, phone, _ := ser.parsePhoneNumber(publishData.SendParams)
	publishData.SendParams.Phone = phone
	publishData.SendParams.Country = country

	record := &sms.SmsRecord{}
	ser.DB.Where("id = ?", publishData.SmsRecordId).First(&record)
	t := time.Now()
	record.LastSendAt = &t
	for _, provider := range ser.Providers {
		if ser.validProvider(provider, publishData.SendParams, publishData.SentProviders, country) {
			if providerResp, externalID, err = provider.Send(*publishData.SendParams); err != nil {
				ser.DB.Save(&sms.SmsFailureRecord{
					SmsRecordId:  publishData.SmsRecordId,
					ProviderName: provider.GetCode(),
					Phone:        phone,
					Error:        err.Error(),
				})
				if record.Error == "" {
					record.Error = fmt.Sprintf("%v: %v;", provider.GetCode(), err)
				} else {
					record.Error = record.Error + fmt.Sprintf("%v: %v;", provider.GetCode(), err)
				}
				record.ExternalID = externalID
				record.Sender = provider.GetCode()
				ser.DB.Save(&record)
				publishData.SentProviders = append(publishData.SentProviders, provider.GetCode())
				ser.Queue.Publish(publishData)
				return
			}
			record.Sender = provider.GetCode()
			record.ExternalID = externalID
			record.ProviderResp = providerResp
			record.State = sms.RecordState_Success
			ser.DB.Save(&record)
			return
		}
	}
	record.State = sms.RecordState_Failure
	ser.DB.Save(&record)
	return
}

func (ser SmsServer) Liveness() ([]byte, bool) {
	var (
		dbALive       = health_checker.PingDB(ser.DB)
		rabbitMQALive = ser.Queue.Liveness()
		err           error
		data          []byte
		alive         = false
	)
	healthCheck := struct {
		DBALive       bool
		RabbitMQALive bool
	}{
		DBALive: dbALive, RabbitMQALive: rabbitMQALive,
	}
	if dbALive && rabbitMQALive {
		alive = true
	}
	data, err = json.Marshal(healthCheck)
	if err != nil {
		return []byte(""), false
	}
	return data, alive
}

func (ser SmsServer) Readiness() bool {
	return true
}

func (ser SmsServer) parsePhoneNumber(params *sms.SendParams) (country string, number string, err error) {
	var (
		countryCode = params.Country
		phoneNumber *libphonenumber.PhoneNumber
	)
	if countryCode == "" {
		countryCode = libphonenumber.UNKNOWN_REGION
	}
	if phoneNumber, err = libphonenumber.Parse(params.Phone, countryCode); err == nil {
		return libphonenumber.GetRegionCodeForCountryCode(int(phoneNumber.GetCountryCode())), libphonenumber.Format(phoneNumber, libphonenumber.INTERNATIONAL), nil
	}
	if phoneNumber, err = libphonenumber.Parse("+"+params.Phone, countryCode); err == nil {
		return libphonenumber.GetRegionCodeForCountryCode(int(phoneNumber.GetCountryCode())), libphonenumber.Format(phoneNumber, libphonenumber.INTERNATIONAL), nil
	}
	return "", "", err
}

func (ser SmsServer) validProvider(provider sms.SmsProvider, sendParams *sms.SendParams, previousProvider []string, country string) bool {
	for _, v := range previousProvider {
		if v == provider.GetCode() {
			return false
		}
	}
	if len(provider.AvailableCountries()) == 0 && provider.Available(sendParams) {
		return true
	}
	for _, code := range provider.AvailableCountries() {
		if code == country && provider.Available(sendParams) {
			return true
		}
	}
	return false
}
