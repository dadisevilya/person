package gettOps

import (
	"strings"

	"github.com/newrelic/go-agent"
	"github.com/gtforge/gls"
)

func InstrumentDatastore(name, collection string, block func() error) error {
	if txn := gls.Get(GlsNewRelicTxnKey); txn != nil {
		tx := txn.(newrelic.Transaction)
		segment := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: strings.ToUpper(collection),
			Operation:  strings.TrimSpace(name),
			StartTime:  newrelic.StartSegmentNow(tx),
		}

		defer segment.End()
	}

	return block()
}

func InstrumentBlock(name string, block func()) {
	if txn := gls.Get(GlsNewRelicTxnKey); txn != nil {
		tx := txn.(newrelic.Transaction)
		segment := newrelic.StartSegment(tx, name)
		defer segment.End()
	} else {
		tx := Newrelic.StartTransaction(name, nil, nil)
		gls.Set(GlsNewRelicTxnKey, txn)
		defer gls.Cleanup()
		defer tx.End()
	}

	block()
}
