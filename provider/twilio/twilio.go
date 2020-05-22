package twilio

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/subosito/twilio"
	"github.com/tidwall/gjson"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
)

type TwilioProvider struct {
	DB *gorm.DB
}

func New() TwilioProvider {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	return TwilioProvider{DB: db}
}

func (provider TwilioProvider) GetCode() string {
	return "Twilio"
}

func (provider TwilioProvider) AvailableCountries() []string {
	return []string{}
}

func (provider TwilioProvider) Available(s *sms.SendParams) bool {
	b := sms.SmsBrand{}
	provider.DB.Where("name = ?", s.Brand).First(&b)
	if !b.EnableTwilio || b.TwilioAccountsID == "" || b.TwilioAuthToken == "" || b.TwilioSendNumber == "" {
		return false
	}
	return true
}

func (provider TwilioProvider) Send(params sms.SendParams) (string, string, error) {
	c := config.MustGetConfig()
	brand := sms.SmsBrand{}
	provider.DB.Where("name = ?", params.Brand).First(&brand)
	SharedClient := twilio.NewClient(brand.TwilioAccountsID, brand.TwilioAuthToken, nil)
	message, _, err := SharedClient.Messages.Send(brand.TwilioSendNumber, params.Phone, twilio.MessageParams{
		Body:           params.Content,
		StatusCallback: c.TwilioCallBack,
	})
	if err != nil {
		return "", "", err
	}
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", "", err
	}
	return string(jsonData), gjson.Get(string(jsonData), "sid").String(), err
}
