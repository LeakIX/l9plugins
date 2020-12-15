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

func (plugin MongoOpenPlugin) Run(ctx context.Context, event *l9format.L9Event) (leak l9format.L9LeakEvent, hasLeak bool) {
	leak.Severity = l9format.SEVERITY_HIGH
	leak.Type = "open_database"
	leak.Data = ""
	hasLeak = false
	//Build client
	deadline, hasDeadline := ctx.Deadline()
	mongoUrl := fmt.Sprintf("mongodb://%s", net.JoinHostPort(event.Ip, event.Port))
	cs, err := connstring.ParseAndValidate(mongoUrl)
	if err != nil {
		log.Println(err)
		return leak, hasLeak
	}
	if hasDeadline {
		cs.ConnectTimeout = deadline.Sub(time.Now())
	}
	tp, err := topology.New(topology.WithConnString(func(connstring.ConnString) connstring.ConnString { return cs }))
	if err != nil {
		log.Println(err)
		return leak, hasLeak
	}
	err = tp.Connect()
	if err != nil {
		log.Println(err)
		return leak, hasLeak
	}
	defer tp.Disconnect(ctx)
	// List collections
	op := operation.NewListCollections(nil).Deployment(tp).Database("admin")
	err = op.Execute(ctx)
	if err != nil {
		log.Println(err)
		return leak, hasLeak
	}
	cursor, err := op.Result(driver.CursorOptions{BatchSize: 20})
	if err != nil {
		log.Println(err)
		return leak, hasLeak
	}
	for cursor.Next(ctx) {
		documents, err := cursor.Batch().Documents()
		if err != nil {
			return leak, hasLeak
		}
		for _, document := range documents {
			leak.Dataset.Collections++
			leak.Data += fmt.Sprintf("Found collection %s\n", document.Lookup("name").String())
		}
		if leak.Dataset.Collections > 128 {
			break
		}
	}
	if leak.Dataset.Collections > 0 {
		hasLeak = true
		leak.Data = fmt.Sprintf("Found %d collections:\n%s", leak.Dataset.Collections, leak.Data)
	}
	return leak, hasLeak
}
