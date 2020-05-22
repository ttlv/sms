package testingtool

import (
	"fmt"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/theplant/testingutils"
)

type Record struct {
	Value string
}

/*
  解析:
      Name: OCX;     Domains: www.ocx.com
      Name: Laffles; Domains: www.laffles.com
      Name: OCXJP;   Domains: www.ocx.co.jp;   ParentExchangeID: 1
  为key, value的数组用于post创建时的参数
*/
func ParseParams(dataStr string, prefix ...string) (results []map[string]string) {
	compactDatas := strings.Replace(dataStr, "\t", "", -1)
	for _, data := range strings.Split(compactDatas, "\n") {
		data = strings.TrimSpace(data)
		if data != "" && !strings.HasPrefix(data, "//") {
			params := make(map[string]string)
			for _, attr := range strings.Split(data, ";") {
				kv := strings.Split(attr, ":")
				if len(prefix) > 0 {
					params[fmt.Sprintf("%v%v", prefix[0], strings.TrimSpace(kv[0]))] = strings.TrimSpace(kv[1])
				} else {
					params[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
				}
			}
			results = append(results, params)
		}
	}
	return
}

/*
  解析:
				  Filters |         Fields | ExpectedResults
				UserID: 1 |       MemberID | 1; 2; 3
  为[{
	  "Filters": ["UserID: 1"],
	  "Fields": [],
	  "ExpectedResults": ["1", "2", "3"]
    }]
*/
func ParseParamsWithHeader(dataStr string) (results []map[string][]string) {
	compactDatas := strings.Replace(dataStr, "\t", "", -1)
	datas := []string{}
	for _, data := range strings.Split(compactDatas, "\n") {
		data = strings.TrimSpace(data)
		if data != "" && !strings.HasPrefix(data, "//") {
			datas = append(datas, data)
		}
	}
	headers := []string{}
	for _, header := range strings.Split(datas[0], "|") {
		headers = append(headers, strings.TrimSpace(header))
	}
	for _, data := range datas[1:len(datas)] {
		params := make(map[string][]string)
		for i, attr := range strings.Split(data, "|") {
			formattedValues := []string{}
			values := strings.Split(strings.TrimSpace(attr), ";")
			for _, value := range values {
				formattedValues = append(formattedValues, strings.TrimSpace(value))
			}
			params[headers[i]] = formattedValues
		}
		results = append(results, params)
	}
	return
}

/*
  对比
	1,ocx,100; 2,eth,100
  和
  `
	1,ocx,100
	2,eth,100
  `
*/

func CompareRecords(t *testing.T, expected string, got string) {
	var (
		expectedValues = []string{}
		gotValues      = []string{}
	)
	for _, line := range strings.Split(expected, "\n") {
		line = strings.TrimSpace(line)
		line = strings.Replace(line, "\t", "", -1)
		if line != "" {
			expectedValues = append(expectedValues, line)
		}
	}
	for _, line := range strings.Split(got, ";") {
		if strings.TrimSpace(line) != "" {
			gotValues = append(gotValues, strings.TrimSpace(line))
		}
	}
	if diff := testingutils.PrettyJsonDiff(expectedValues, gotValues); len(diff) > 0 {
		t.Errorf("CompareRecords:\n %v", diff)
	}
}

func GetRecords(db *gorm.DB, tableName string, columns string, extra ...string) string {
	var (
		extraSQL = ""
		results  = []string{}
		records  = []Record{}
	)
	if len(extra) > 0 {
		extraSQL = extra[0]
	}
	db.Raw(fmt.Sprintf(`SELECT CONCAT_WS(',', %v) AS value FROM %v %v`, columns, tableName, extraSQL)).Scan(&records)
	for _, record := range records {
		results = append(results, record.Value)
	}
	return strings.Join(results, "; ")
}
