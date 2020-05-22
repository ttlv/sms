package admin

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	qor_admin "github.com/qor/admin"
	"github.com/ttacon/libphonenumber"
	"github.com/ttlv/sms"
)

func parsePhoneNumber(phone string) (country string, number string, err error) {
	var phoneNumber *libphonenumber.PhoneNumber
	if phoneNumber, err = libphonenumber.Parse(phone, ""); err == nil {
		return libphonenumber.GetRegionCodeForCountryCode(int(phoneNumber.GetCountryCode())), libphonenumber.Format(phoneNumber, libphonenumber.INTERNATIONAL), nil
	}
	if phoneNumber, err = libphonenumber.Parse("+"+phone, ""); err == nil {
		return libphonenumber.GetRegionCodeForCountryCode(int(phoneNumber.GetCountryCode())), libphonenumber.Format(phoneNumber, libphonenumber.INTERNATIONAL), nil
	}
	return "", "", err
}

func constructParam(actionArgument *qor_admin.ActionArgument) *sms.SendParams {
	country, phone, _ := parsePhoneNumber(actionArgument.Argument.(*SendParamForm).Phone)
	params := sms.SendParams{
		Country: country,
		Brand:   actionArgument.FindSelectedRecords()[0].(*sms.SmsBrand).Name,
		Phone:   phone,
		Content: actionArgument.Argument.(*SendParamForm).Content,
	}
	return &params
}

func saveSmsRecord(DB *gorm.DB, param *sms.SendParams, sender string, resp string, externalID string) {
	rawParamByte, _ := json.Marshal(param)
	smsRecord := sms.SmsRecord{
		Brand:        param.Brand,
		Phone:        param.Phone,
		RawParam:     string(rawParamByte),
		State:        sms.RecordState_Success,
		Sender:       sender,
		ProviderResp: resp,
		ExternalID:   externalID,
	}
	DB.Save(&smsRecord)
}
