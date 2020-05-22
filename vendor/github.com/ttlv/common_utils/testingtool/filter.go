package testingtool

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	theplant_test "github.com/theplant/testingutils"
)

type FilterTestCase struct {
	URL   string
	Cases string
}

func RunFilterCase(t *testing.T, server *httptest.Server, testCase FilterTestCase, allFilter []string) {
	color.Green("RunFilterCase: %v", testCase.URL)
	datas := ParseParamsWithHeader(testCase.Cases)
	for _, data := range datas {
		filters := []string{}
		tempMap := map[string]bool{}
		if len(data["Filters"]) != 0 && data["Filters"][0] != "" {
			for _, filter := range data["Filters"] {
				location := strings.IndexAny(filter, ":")
				var kv [2]string
				kv[0] = strings.TrimSpace(filter[0:location])
				tempMap[kv[0]] = true
				kv[1] = strings.Replace(strings.TrimSpace(filter[location+1:]), " ", "%20", -1)
				if strings.Contains(kv[0], "filters") {
					filters = append(filters, strings.TrimSpace(kv[0])+"="+strings.TrimSpace(kv[1]))
				} else {
					filters = append(filters, "filters%5B"+strings.TrimSpace(kv[0])+"%5D.Value="+strings.TrimSpace(kv[1]))
				}
			}
		}
		// 去重
		for _, value := range allFilter {
			if tempMap[value] {
				continue
			}
			if strings.Contains(value, "filters") {
				filters = append(filters, strings.TrimSpace(value))
			} else {
				filters = append(filters, "filters%5B"+strings.TrimSpace(value)+"%5D.Value=")
			}
		}
		doc, _ := goquery.NewDocument(server.URL + testCase.URL + "?" + strings.Join(filters, "&"))
		results := []string{}
		doc.Find(".qor-table-container tbody tr").Each(func(i int, tr *goquery.Selection) {
			result := []string{}
			tr.Find("td").Each(func(j int, td *goquery.Selection) {
				for _, attr := range data["Fields"] {
					if td.AttrOr("data-heading", "") == attr {
						result = append(result, strings.TrimSpace(td.Text()))
					}
				}
			})
			results = append(results, strings.Join(result, ","))
		})
		if diff := theplant_test.PrettyJsonDiff(data["ExpectedResults"], results); len(diff) > 0 {
			t.Errorf("%v ExpectedResults %v: %v", testCase.URL, data["Filters"][0], diff)
		}
	}
}

/*demo:
          筛选条件:                                                                              需要检查的字段           期望值
		  Filters                                                                       		| Fields                 | ExpectedResults
																				       		 	| CurrencySymbol         | OCX; ETH; BTC
		  filters[Date].Start: 2018-10-11                                             		 	| Date; CurrencySymbol   | 2018-10-12,OCX; 2018-10-11,ETH
		  filters[Date].End: 2018-10-11                                                		 	| Date; CurrencySymbol   | 2018-10-11,ETH; 2018-10-10,BTC
		  filters[Date].Start: 2018-10-10; filters[Date].End: 2018-10-11              		  	| Date; CurrencySymbol   | 2018-10-11,ETH; 2018-10-10,BTC
 		  filters[Email].Value: 2@qq.com														| Email; CurrencySymbol  | 2@qq.com,ETH
		  MemberID: 1																			| CurrencySymbol		 | BTC
		  PhoneNumber: 18600000001                                                      		| CurrencySymbol    	 | BTC
 		  State:0																				| CurrencySymbol		 | OCX; BTC
		  MemberID: 1; filters[Date].Start: 2018-10-10; filters[Date].End: 2018-10-11			| CurrencySymbol; Date   | 2018-10-10,BTC

*/
