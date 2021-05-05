package gettStorages

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"

	gettConfig "github.com/gtforge/global_services_common_go/gett-config"
	gettOps "github.com/gtforge/global_services_common_go/gett-ops"
	tracingHelper "github.com/gtforge/global_services_common_go/gett-ops/opentracing"
	gettSqlDrivers "github.com/gtforge/global_services_common_go/gett-storages/gett-sql-drivers"
	"github.com/gtforge/gls"
	"github.com/gtforge/gorm"
	_ "github.com/lib/pq"
	newrelic "github.com/newrelic/go-agent"
)

var DB *gorm.DB
var dbOnce sync.Once
var noopTracer opentracing.NoopTracer

const (
	defaultPool                     = 30
	newrelicStart                   = "newrelicStart"
	newrelicStop                    = "newrelicStop"
	opentracingStart                = "opentracingStart"
	opentracingStop                 = "opentracingStop"
	gormBeginTransaction            = "gorm:begin_transaction"
	gormCommitOrRollbackTransaction = "gorm:commit_or_rollback_transaction"
	gormQuery                       = "gorm:query"
	gormRowQuery                    = "gorm:row_query"
	gormAfterQuery                  = "gorm:after_query"
	gormAssignUpdatingAttributes    = "gorm:assign_updating_attributes"
	opentracingSpanKey              = "opentracingSpanKey"
)

type DBConf struct {
	UserName         string
	Password         string
	Host             string
	Database         string
	IdlePool         int
	OpenConsPool     int
	DatastorePostfix string // postfix for database in NewRelic, ex: "Replica"
	ReadOnly         bool
	MaxConLifeTime   time.Duration
}

func InitDb(dbConfig *gettConfig.Config, appEnv string, opts ...DBOption) {
	if len(dbConfig.AllKeys()) > 0 {
		dbOnce.Do(func() {
			DB = NewDBClientWithConf(DBConf{
				UserName:       dbConfig.GetString(appEnv + ".user_name"),
				Password:       dbConfig.GetString(appEnv + ".password"),
				Host:           dbConfig.GetString(appEnv + ".host"),
				Database:       dbConfig.GetString(appEnv + ".database"),
				IdlePool:       dbConfig.GetInt(appEnv + ".idle_pool"),
				OpenConsPool:   dbConfig.GetInt(appEnv + ".open_pool"),
				ReadOnly:       dbConfig.GetBool(appEnv + ".read_only"),
				MaxConLifeTime: dbConfig.GetDuration(appEnv + ".max_con_life_time"),
			}, opts...)
		})
	}
}

func NewDBClientWithConf(conf DBConf, opts ...DBOption) *gorm.DB {
	dbUri := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", conf.UserName, conf.Password, conf.Host, conf.Database)
	dbOption := dbOptionsApply(opts...)
	var driverName = "postgres"
	if conf.ReadOnly {
		driverName = postgresRO
	}
	db, err := gorm.Open("postgres", driverName, dbUri)
	if err != nil {
		panic(fmt.Sprintf("Database connection error: %v", err))
	}

	err = db.DB().Ping()
	if err != nil {
		panic(fmt.Sprintf("Database connection error: %v", err))
	}

	idlePool := conf.IdlePool
	if idlePool == 0 {
		idlePool = defaultPool
	}

	openConsPool := conf.OpenConsPool
	if openConsPool == 0 {
		openConsPool = defaultPool
	}

	db.DB().SetMaxIdleConns(idlePool)
	db.DB().SetMaxOpenConns(openConsPool)
	db.DB().SetConnMaxLifetime(conf.MaxConLifeTime)
	db.LogMode(gettConfig.Settings.Db.GetBool(gettConfig.Settings.AppEnv + ".log"))

	dbNewRelicStart := buildDBNewRelicStartCallback(conf)
	dbOpenTracingStart := dbOpenTracingStart(dbOption.tracer)

	registerBeforeCallbacks(db, newrelicStart, dbNewRelicStart)
	registerBeforeCallbacks(db, opentracingStart, dbOpenTracingStart)

	registerAfterCallbacks(db, newrelicStop, dbNewRelicStop)
	registerAfterCallbacks(db, opentracingStop, dbOpenTracingStop)

	return db
}

func registerBeforeCallbacks(db *gorm.DB, callbackName string, callback func(scope *gorm.Scope)) {
	db.Callback().Create().Before(gormBeginTransaction).Register(callbackName, callback)
	db.Callback().Delete().Before(gormBeginTransaction).Register(callbackName, callback)
	db.Callback().Query().Before(gormQuery).Register(callbackName, callback)
	db.Callback().RowQuery().Before(gormRowQuery).Register(callbackName, callback)
	db.Callback().Update().Before(gormAssignUpdatingAttributes).Register(callbackName, callback)
}

func registerAfterCallbacks(db *gorm.DB, callbackName string, callback func(scope *gorm.Scope)) {
	db.Callback().Create().After(gormCommitOrRollbackTransaction).Register(callbackName, callback)
	db.Callback().Delete().After(gormCommitOrRollbackTransaction).Register(callbackName, callback)
	db.Callback().Query().After(gormAfterQuery).Register(callbackName, callback)
	db.Callback().RowQuery().After(gormAfterQuery).Register(callbackName, callback)
	db.Callback().Update().After(gormCommitOrRollbackTransaction).Register(callbackName, callback)
}

func ExecuteInTransaction(db *gorm.DB, txCode func(tx *gorm.DB) error) error {
	var committed bool
	tx := db.Begin()
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	err := txCode(tx)
	if err != nil {
		return err
	}

	committed = true
	return tx.Commit().Error
}

func dbOpenTracingStart(tracer opentracing.Tracer) func(scope *gorm.Scope) {
	return func(scope *gorm.Scope) {
		var parentSpanContext opentracing.SpanContext

		if parent := tracingHelper.SpanFromGLS(); parent != nil {
			parentSpanContext = parent.Context()
		}

		span := tracer.StartSpan(
			"sql.query",
			opentracing.ChildOf(parentSpanContext),
		)
		otext.DBType.Set(span, "sql")
		spanToScope(scope, span)
	}
}

func sqlOperationName(sql string) string {
	return strings.ToUpper(strings.Split(strings.TrimSpace(sql), " ")[0])
}

func spanToScope(scope *gorm.Scope, span opentracing.Span) *gorm.Scope {
	return scope.Set(opentracingSpanKey, span)
}

func spanFromScope(scope *gorm.Scope) opentracing.Span {
	if value, ok := scope.Get(opentracingSpanKey); ok {
		if span, ok := value.(opentracing.Span); ok {
			return span
		}
	}
	return noopTracer.StartSpan("")
}

func dbOpenTracingStop(scope *gorm.Scope) {
	span := spanFromScope(scope)
	span.SetOperationName("sql.query " + sqlOperationName(scope.SQL)) // sql statement is empty in `before` callback
	if scope.HasError() {
		otext.Error.Set(span, true)
		span.SetTag("error.details", scope.DB().Error)
	}
	otext.DBStatement.Set(span, scope.SQL)
	span.SetTag("db.table", scope.TableName())
	span.SetTag("db.statement", scope.SQL)
	span.Finish()
}

func buildDBNewRelicStartCallback(conf DBConf) func(*gorm.Scope) {
	var productName newrelic.DatastoreProduct
	if conf.DatastorePostfix == "" {
		productName = newrelic.DatastorePostgres
	} else {
		productName = newrelic.DatastoreProduct(fmt.Sprintf("%s %s", newrelic.DatastorePostgres, conf.DatastorePostfix))
	}

	return func(scope *gorm.Scope) {
		if content := gls.Get(gettOps.GlsNewRelicTxnKey); content != nil {
			txn := content.(newrelic.Transaction)
			s := newrelic.DatastoreSegment{
				Product:    productName,
				Collection: strings.ToUpper(scope.TableName()),
				Operation:  sqlOperationName(scope.SQL),
				StartTime:  newrelic.StartSegmentNow(txn),
			}
			gls.Set("newrelic_db_txn", s)
		}
	}
}

func dbNewRelicStop(scope *gorm.Scope) {
	if content := gls.Get("newrelic_db_txn"); content != nil {
		segment := content.(newrelic.DatastoreSegment)
		segment.Operation = sqlOperationName(scope.SQL)
		segment.ParameterizedQuery = scope.SQL

		segment.End()
	}
}

var postgresRO = gettSqlDrivers.NewDriver(
	"global_postgres",
	&driver.TxOptions{ReadOnly: true},
)
