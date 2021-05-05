package transactionid

import (
	"github.com/gtforge/global_services_common_go/gett-utils"
	"github.com/gtforge/gls"
)

const RequestIdKey = "X-Request-Id"

func GetTransactionId() string {
	tx := gls.Get(RequestIdKey)
	if tx == nil {
		tx, _ = gettUtils.NewUUID()
		gls.Set(RequestIdKey, tx)
	}

	return tx.(string)
}
