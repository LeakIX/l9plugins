package tcp

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

func (plugin MysqlSchemaPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) ( hasLeak bool) {
	if len(event.Service.Credentials.Username) < 1 {
		log.Printf("No credentials found for %s:%s", net.JoinHostPort(event.Host, event.Port))
		return false
	}
	dsn := fmt.Sprintf("%s@l9tcp(%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", event.Service.Credentials.Username, net.JoinHostPort(event.Ip, event.Port))
	if len(event.Service.Credentials.Password) > 0 {
		dsn = fmt.Sprintf("%s:%s@l9tcp(%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", event.Service.Credentials.Username, event.Service.Credentials.Password, net.JoinHostPort(event.Ip, event.Port))
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return false
	}
	defer db.Close()
	var databaseName string
	var tableName string
	var recordCount int64
	var dataLength int64
	tableListQuery, err := db.QueryContext(ctx,"SELECT TABLE_SCHEMA, TABLE_NAME, TABLE_ROWS, DATA_LENGTH from  TABLES where table_schema != 'information_schema' AND table_schema != 'sys' AND table_schema != 'performance_schema';")
	if err != nil {
		log.Println(err.Error())
		return hasLeak
	}

	for tableListQuery.Next() {
		tableListQuery.Scan(&databaseName, &tableName, &recordCount, &dataLength)
		log.Printf("Found table %s.%s with %d records\n", databaseName, tableName, recordCount)
		event.Summary += fmt.Sprintf("Found table %s.%s with %d records\n", databaseName, tableName, recordCount)
		event.Leak.Dataset.Rows += recordCount
		event.Leak.Dataset.Collections++
		event.Leak.Dataset.Size += dataLength
		if strings.Contains(strings.ToLower(tableName), "warning") || strings.HasPrefix(strings.ToLower(tableName), "readme_") || strings.HasPrefix(strings.ToLower(tableName), "recover_your_"){
			event.Leak.Dataset.Infected = true
			if ransomNote, found := plugin.GetRansomNote(ctx, databaseName, tableName ,event, db) ; found{
				event.Leak.Dataset.RansomNotes = append(event.Leak.Dataset.RansomNotes, ransomNote)
			}
		}
	}
	if event.Leak.Dataset.Collections < 1 {
		return false
	}
	event.Summary = fmt.Sprintf("Databases: %d, row count: %d, size: %s\n",
		event.Leak.Dataset.Collections, event.Leak.Dataset.Rows, utils.HumanByteCount(event.Leak.Dataset.Size)) +
		event.Summary
	event.Leak.Severity = l9format.SEVERITY_MEDIUM
	if event.Leak.Dataset.Infected {
		event.Leak.Severity = l9format.SEVERITY_HIGH
	}
	if event.Leak.Dataset.Rows > 1000 {
		event.Leak.Severity = l9format.SEVERITY_HIGH
		if event.Leak.Dataset.Infected {
			event.Leak.Severity = l9format.SEVERITY_CRITICAL
		}
	}
	return true
}

func (MysqlSchemaPlugin) GetRansomNote(ctx context.Context, databaseName, tableName string, event *l9format.L9Event, db *sql.DB) (ransomNote string, found bool) {
	queryResult, err := db.Query(fmt.Sprintf("SELECT * FROM `%s`.`%s` LIMIT 1", databaseName, tableName))
	if err != nil {
		log.Println(err)
		return "", false
	}
	queryResult.Next()
	columnNames, _ := queryResult.Columns()
	results := make([]interface{}, len(columnNames))
	for idx, _ := range results {
		theString := ""
		results[idx] = &theString
	}
	var note string
	err = queryResult.Scan(results...)
	if err != nil {
		log.Println(err)
		return "", false
	}
	for _, columnContent := range results {
		if columnText, isString := columnContent.(*string); isString && len(*columnText) > 0{
			note += *columnText
		}
	}
	if len(note) > 1 {
		return note, true
	}
	return "", false
}

var mysqlRegistered bool

func (plugin MysqlSchemaPlugin) Init() error {
	mysql.RegisterDialContext("l9tcp", func(ctx context.Context, remoteAddr string) (net.Conn, error) {
		return plugin.DialContext(ctx, "tcp", remoteAddr)
	})
	log.Println("Registered l9tcp mysql dialer")
	return nil
}
