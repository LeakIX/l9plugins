package main

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/connstring"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
	"log"
	"net"
	"time"
)

type MongoOpenPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return MongoOpenPlugin{}
}

// Implement interface :
func (MongoOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}
func (MongoOpenPlugin) GetProtocols() []string {
	return []string{"mongo"}
}
func (MongoOpenPlugin) GetName() string {
	return "MongoOpenPlugin"
}

func (MongoOpenPlugin) GetStage() string {
	return l9format.STAGE_OPEN
}

func (plugin MongoOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) (hasLeak bool) {
	event.Leak.Severity = l9format.SEVERITY_HIGH
	event.Leak.Type = "open_database"
	//Build client
	deadline, hasDeadline := ctx.Deadline()
	mongoUrl := fmt.Sprintf("mongodb://%s/", net.JoinHostPort(event.Ip, event.Port))
	if event.HasTransport("tls") {
		mongoUrl += "?tls=true&tlsAllowInvalidCertificates=true&tlsInsecure=true"
	}
	cs, err := connstring.ParseAndValidate(mongoUrl)
	if err != nil {
		log.Println(err)
		event.Summary += err.Error() + "\n"
		return false
	}
	if hasDeadline {
		cs.ConnectTimeout = deadline.Sub(time.Now())
	}
	tp, err := topology.New(topology.WithConnString(func(connstring.ConnString) connstring.ConnString { return cs }))
	if err != nil {
		log.Println(err)
		return false
	}
	err = tp.Connect()
	if err != nil {
		log.Println(err)
		event.Summary += err.Error() + "\n"
		return false
	}
	defer tp.Disconnect(ctx)
	// List collections
	op := operation.NewListCollections(nil).Deployment(tp).Database("admin")
	err = op.Execute(ctx)
	if err != nil {
		log.Println(err)
		event.Summary += err.Error() + "\n"
		return false
	}
	cursor, err := op.Result(driver.CursorOptions{BatchSize: 20})
	if err != nil {
		log.Println(err)
		event.Summary += err.Error() + "\n"
		return false
	}
	for cursor.Next(ctx) {
		documents, err := cursor.Batch().Documents()
		if err != nil {
			event.Summary += err.Error() + "\n"
			return false
		}
		for _, document := range documents {
			event.Leak.Dataset.Collections++
			event.Summary += fmt.Sprintf("Found collection %s\n", document.Lookup("name").String())
		}
		if event.Leak.Dataset.Collections > 128 {
			break
		}
	}
	if event.Leak.Dataset.Collections > 0 {
		event.Summary = fmt.Sprintf("Found %d collections:\n%s", event.Leak.Dataset.Collections, event.Summary)
		return true
	}
	return false
}
