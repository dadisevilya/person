package gettOps

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gtforge/global_services_common_go/gett-utils/transactionid"

	"github.com/sirupsen/logrus"
)

const timestampFormat = time.RFC3339

type GettLogFormatter struct{}

func (f *GettLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	data := make(map[string]interface{}, len(entry.Data)+4)
	data["level"] = entry.Level.String()
	data["time"] = entry.Time.UTC().Format(timestampFormat)
	data["requestId"] = transactionid.GetTransactionId()

	for k, v := range entry.Data {
		switch v := v.(type) {
		case error:
			// Otherwise errors are ignored by `encoding/json`
			data[k] = v.Error()
		default:
			data[k] = v
		}
	}

	data["message"] = entry.Message

	serialized, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %v", err)
	}
	return append(serialized, '\n'), nil
}
