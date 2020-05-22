package yunpian

import (
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
)

func TestYunPian(t *testing.T) {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&sms.SmsBrand{})
	utils.RunSQL(db, `
	  TRUNCATE table sms_brands;
	  INSERT INTO sms_brands (name, yun_pian_app_key, yun_pian_host) VALUES ('xxx', 'your yun pian app key', 'your yun pian host');
	`)

	provider := New()
	resp, externalID, err := provider.Send(sms.SendParams{
		Phone:   "",
		Country: "CN",
		Content: fmt.Sprintf("【xxx】您的验证码是%v。如非本人操作，请忽略本短信", time.Now().Unix()),
		Brand:   "xxx",
	})
	fmt.Printf("Err: %v\n", err)
	fmt.Printf("返回的结果: %v\n", resp)
	fmt.Printf("返回的ExternalID: %v\n", externalID)
}
