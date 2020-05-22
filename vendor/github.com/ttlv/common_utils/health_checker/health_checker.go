package health_checker

import (
	"log"
	"net/http"

	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

type HeathChecker interface {
	Liveness() ([]byte, bool)
	Readiness() bool
}

func PingDB(db *gorm.DB) bool {
	if err := db.Exec("SHOW TABLES"); err.Error != nil {
		return false
	}
	return true
}

func PingRedis(client *redis.Client) bool {
	if _, err := client.Ping().Result(); err != nil {
		return false
	}
	return true
}

func StartCheck(c HeathChecker) {
	go func(checker HeathChecker) {
		http.HandleFunc("/health_check/liveness", func(w http.ResponseWriter, r *http.Request) {
			data, result := checker.Liveness()
			if result {
				w.WriteHeader(http.StatusOK)
				w.Write(data)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(data)
			return
		})
		http.HandleFunc("/health_check/readiness", func(w http.ResponseWriter, r *http.Request) {
			_, result := checker.Liveness()
			if result {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("ok"))
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
			return
		})
		log.Fatal(http.ListenAndServe(":9999", nil))
	}(c)
}
