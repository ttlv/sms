package amazon

import (
	"fmt"
	"testing"
	"time"

	"github.com/jinzhu/gorm"

	"github.com/ttlv/common_utils/utils"
	"github.com/ttlv/sms"
	"github.com/ttlv/sms/config"
)

func TestAmazon(t *testing.T) {
	cfg := config.MustGetConfig()
	db, err := gorm.Open("mysql", cfg.DB)
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&sms.SmsBrand{})
	utils.RunSQL(db, `
	  TRUNCATE table sms_brands;
	  INSERT INTO sms_brands (name, aws_access_key_id, aws_secret_access_key, aws_region) VALUES ('xxx', 'your aws access key id', 'your aws secret access key', 'your aws region');
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
