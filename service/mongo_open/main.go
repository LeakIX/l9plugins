package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/LeakIX/l9format"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
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
	hasLeak = false
	deadline, hasDeadline := ctx.Deadline()

	mongoUrl := fmt.Sprintf("mongodb://%s", net.JoinHostPort(event.Ip, event.Port))

	connectOptions := options.Client().
		ApplyURI(mongoUrl).
		SetDialer(plugin)
	if event.HasTransport("tls") {
		connectOptions.SetTLSConfig(&tls.Config{InsecureSkipVerify: true})
	}
	connectOptions.SetDisableOCSPEndpointCheck(true)
	if hasDeadline {
		connectOptions.SetConnectTimeout(deadline.Sub(time.Now()))
	}
	client, err := mongo.Connect(ctx, connectOptions)
	if err != nil {
		log.Println("Connect error: " + err.Error())
		return leak, hasLeak
	}
	err = client.Ping(ctx, reeadpref.)
	defer client.Disconnect(ctx)
	dbList, err := client.ListDatabases(nil, bson.D{{}})
	if err != nil {
		log.Println("ListDB error: " + err.Error())
		return leak, hasLeak
	}
	event.Service.Credentials = l9format.ServiceCredentials{
		NoAuth: true,
	}
	// Fill some server info
	event.Service.Software.OperatingSystem, event.Service.Software.Version, _ = getServerVersion(ctx, client)
	event.Service.Software.Name = "MongoDB"
	if len(dbList.Databases) > 0 {
		hasLeak = true
		leak.Type = "open_database"
		leak.Severity = l9format.SEVERITY_HIGH
		leak.Data = fmt.Sprintf("No authentication on MongoDB server, found %d collections", len(dbList.Databases))
		leak.Dataset.Collections = int64(len(dbList.Databases))
	}
	return leak, hasLeak
}

func getServerVersion(ctx context.Context, client *mongo.Client) (string, string, error) {
	serverStatus, err := client.Database("admin").RunCommand(
		ctx,
		bsonx.Doc{{"serverStatus", bsonx.Int32(1)}},
	).DecodeBytes()
	if err != nil {
		return "", "", err
	}
	log.Println("got status")
	version, err := serverStatus.LookupErr("version")
	if err != nil {
		return "", "", err
	}

	hostInfo, err := client.Database("admin").RunCommand(
		ctx,
		bsonx.Doc{{"hostInfo", bsonx.Int32(1)}},
	).DecodeBytes()
	if err != nil {
		return "", version.StringValue(), err
	}
	log.Println("got info")
	os, err := hostInfo.LookupErr("extra", "versionString")
	if err != nil {
		return "", version.StringValue(), err
	}
	log.Println("got extra")
	return os.StringValue(), version.StringValue(), nil
}
