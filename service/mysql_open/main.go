package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"net"
)

type MysqlWeakPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return MysqlWeakPlugin{}
}

func (MysqlWeakPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (MysqlWeakPlugin) GetProtocols() []string {
	return []string{"mysql"}
}

func (MysqlWeakPlugin) GetName() string {
	return "MysqlWeakPlugin"
}

func (MysqlWeakPlugin) GetStage() string {
	return "open"
}

var verQueryString = "select @@version_comment, @@version, concat(@@version_compile_os, \" \", @@version_compile_machine);"

func (plugin MysqlWeakPlugin) Run(ctx context.Context, event *l9format.L9Event) (leak l9format.L9LeakEvent, hasLeak bool) {
	addr := net.JoinHostPort(event.Ip, event.Port)
	mysql.RegisterDialContext("l9tcp", func(_ context.Context, _ string) (net.Conn, error) {
		return plugin.DialContext(ctx, "tcp", addr)
	})
	for _, username := range usernames {
		for _, password := range passwords {
			dsn := fmt.Sprintf("%s:%s@l9tcp(%s:%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", username, password, event.Ip, event.Port )
			log.Printf("Trying: %s", dsn)
			db, err := sql.Open("mysql", dsn )
			err = db.Ping()
			if err != nil {
				db.Close()
				if _, isMysqlError := err.(*mysql.MySQLError); !isMysqlError {
					log.Println(err.Error())
					log.Println("Not a mysql error, leaving early")
					return leak, hasLeak
				}
				continue
			}
			// Try to populate info for the service
			verQuery, err := db.Query(verQueryString)
			if err == nil {
				if verQuery.Next() {
					verQuery.Scan(&event.Service.Software.Name, &event.Service.Software.Version, &event.Service.Software.OperatingSystem)
				}
			}
			db.Close()
			log.Println("Mysql authed, default password")
			event.Service.Credentials = l9format.ServiceCredentials{
				NoAuth:   false,
				Username: username,
				Password: password,
			}
			leak.Data = "No or default MySQL authentication found."
			hasLeak = true
			return leak, hasLeak
		}
	}
	return leak, hasLeak
}

var usernames = []string{
	"root",
}

var passwords = []string{
	"",
	"root",
	"toor",
	"t00r",
	"r00t",
	"mysql",
	"sql",
	"123456",
	"admin",
}
