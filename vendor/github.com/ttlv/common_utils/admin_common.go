package common_utils

import (
	"fmt"
	"reflect"

	"github.com/qor/admin"
	"github.com/qor/qor"
)

type Count struct {
	Value int64
}

func ReplaceFindManyHandler(res *admin.Resource, tableName string) {
	defaultFindManyHandler := res.FindManyHandler
	res.FindManyHandler = func(i interface{}, c *qor.Context) error {
		if _, ok := i.(*int); ok {
			count := Count{}
			res.GetAdmin().DB.Raw(fmt.Sprintf(`SELECT TABLE_ROWS AS value FROM information_schema.TABLES WHERE table_name = '%v'`, tableName)).Scan(&count)
			reflect.Indirect(reflect.ValueOf(i)).SetInt(count.Value)
			return nil
		}
		return defaultFindManyHandler(i, c)
	}
}
