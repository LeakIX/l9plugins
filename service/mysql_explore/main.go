package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/LeakIX/l9format/utils"
	"github.com/go-sql-driver/mysql"
	"log"
	"net"
	"strings"
)

type MysqlSchemaPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	plugin := MysqlSchemaPlugin{}
	mysql.RegisterDialContext("l9tcp", func(ctx context.Context, remoteAddr string) (net.Conn, error) {
		return plugin.DialContext(ctx, "tcp", remoteAddr)
	})
	return plugin
}

func (MysqlSchemaPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (MysqlSchemaPlugin) GetProtocols() []string {
	return []string{"mysql"}
}

func (MysqlSchemaPlugin) GetName() string {
	return "MysqlSchemaPlugin"
}

func (MysqlSchemaPlugin) GetStage() string {
	return "explore"
}

func (plugin MysqlSchemaPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	if len(event.Service.Credentials.Username) < 1 {
		log.Printf("No credentials found for %s:%s", net.JoinHostPort(event.Host, event.Port))
		return leak, false
	}
	dsn := fmt.Sprintf("%s@l9tcp(%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", event.Service.Credentials.Username, net.JoinHostPort(event.Ip, event.Port))
	if len(event.Service.Credentials.Password) > 0 {
		dsn = fmt.Sprintf("%s:%s@l9tcp(%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", event.Service.Credentials.Username, event.Service.Credentials.Password, net.JoinHostPort(event.Ip, event.Port))
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return leak, false
	}
	var databaseName string
	var tableName string
	var recordCount int64
	var dataLength int64
	tableListQuery, err := db.Query("SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_ROWS, DATA_LENGTH from  TABLES where table_schema != 'information_schema' AND table_schema != 'sys' AND table_schema != 'performance_schema';")
	if err != nil {
		log.Println(err.Error())
		return leak, hasLeak
	}

	for tableListQuery.Next() {
		tableListQuery.Scan(&databaseName, &tableName, &recordCount, &dataLength)
		log.Printf("Found table "+databaseName+"."+tableName+" with %d records\n", recordCount)
		leak.Data += fmt.Sprintf("Found table "+databaseName+"."+tableName+" with %d records\n", recordCount)
		leak.Dataset.Rows += recordCount
		leak.Dataset.Collections++
		leak.Dataset.Size += dataLength
		if strings.Contains(strings.ToLower(tableName), "warning") {
			leak.Dataset.Infected = true
		}
	}
	if leak.Dataset.Collections < 1 {
		return leak, false
	}
	leak.Data = fmt.Sprintf("Databases: %d, row count: %d, size: %s\n",
		leak.Dataset.Collections, leak.Dataset.Rows, utils.HumanByteCount(leak.Dataset.Size)) +
		leak.Data
	return leak, true
}
