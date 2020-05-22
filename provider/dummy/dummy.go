package dummy

import (
	"github.com/jinzhu/gorm"
	"github.com/ttlv/sms"
)

type DummyProvider struct {
	DB        *gorm.DB
	Code      string
	Countries []string
	SendFunc  func(sms.SendParams) (string, string, error)
}

func (provider DummyProvider) GetCode() string {
	return provider.Code
}

func (provider DummyProvider) AvailableCountries() []string {
	return provider.Countries
}

func (provider DummyProvider) Available(s *sms.SendParams) bool {
	brand := &sms.SmsBrand{}
	provider.DB.First(&brand, "name = ?", s.Brand)
	switch provider.Code {
	case "YunPian":
		return brand.EnableYunPian
	case "Twilio":
		return brand.EnableTwilio
	case "Amazon":
		return brand.EnableAWS
	case "Emay":
		return brand.EnableEmay
	}
	return false
}

func (provider DummyProvider) Send(params sms.SendParams) (string, string, error) {
	return provider.SendFunc(params)
}
