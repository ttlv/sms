package admin_test

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/ttlv/sms"
	sms_admin "github.com/ttlv/sms/admin"
	"github.com/ttlv/sms/config"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var (
	db     *gorm.DB
	err    error
	Server *httptest.Server
)

func setup() {
	cfg := config.MustGetConfig()
	if db, err = gorm.Open("mysql", cfg.DB); err != nil {
		panic(err)
	}
	db.AutoMigrate(&sms.SmsRecord{}, &sms.SmsFailureRecord{}, &sms.SmsBrand{}, &sms.SmsSetting{})
	db.LogMode(true)
	adm := sms_admin.New()
	mux := http.NewServeMux()
	Server = httptest.NewServer(mux)
	adm.MountTo("/admin", mux)
	if os.Getenv("MODE") == "server" {
		fmt.Printf("Test Server URL: %v\n", Server.URL+"/admin")
		time.Sleep(time.Second * 3000)
	}
}

func TestAdmin(t *testing.T) {
	setup()
}
