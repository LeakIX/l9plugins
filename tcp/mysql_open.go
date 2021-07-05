package tcp

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

func (plugin MysqlWeakPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	for _, username := range usernames {
		for _, password := range passwords {
			dsn := fmt.Sprintf("%s:%s@l9tcp(%s)/information_schema?readTimeout=3s&timeout=3s&writeTimeout=3s", username, password, net.JoinHostPort(event.Ip, event.Port))
			log.Printf("Trying: %s", dsn)
			db, err := sql.Open("mysql", dsn)

			err = db.PingContext(ctx)
			if err != nil {
				db.Close()
				if _, isMysqlError := err.(*mysql.MySQLError); !isMysqlError {
					log.Println(err.Error())
					log.Println("Not a mysql error, leaving early")
					return false
				}
				continue
			}
			// Try to populate info for the service
			verQuery, err := db.QueryContext(ctx, verQueryString)
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
			event.Summary = "No or default MySQL authentication found."
			return true
		}
	}
	return false
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


func (plugin MysqlWeakPlugin) Init() {
	mysql.RegisterDialContext("l9tcp", func(ctx context.Context, remoteAddr string) (net.Conn, error) {
		return plugin.DialContext(ctx, "tcp", remoteAddr)
	})
	log.Println("Registered l9tcp mysql dialer")
}
