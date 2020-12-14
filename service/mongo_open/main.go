package main

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/LeakIX/l9plugins"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"log"
)

type MongoOpenPlugin struct {
	l9plugins.ServicePluginBase
}

func New() l9plugins.ServicePluginInterface {
	return MongoOpenPlugin{}
}

func (MongoOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (MongoOpenPlugin) GetProtocols() []string {
	return []string{"mongo"}
}

func (MongoOpenPlugin) GetName() string {
	return "MongoOpenPlugin"
}

func (MongoOpenPlugin) GetType() string {
	return "open"
}

func (plugin MongoOpenPlugin) Run(ctx context.Context, event *l9format.L9Event) (leak l9format.L9LeakEvent, hasLeak bool) {
	hasLeak = false
	log.Printf("Trying mongodb://%s:%s", event.Ip, event.Port)
	client, err := mongo.Connect(
		ctx, options.Client().ApplyURI(
			fmt.Sprintf("mongodb://%s:%s", event.Ip, event.Port)).SetDialer(
			plugin))
	if err != nil {
		log.Println("Connect error: " + err.Error())
		return leak, hasLeak
	}
	dbList, err := client.ListDatabases(ctx, bson.D{{}})
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
		leak.Type = "database"
		leak.Severity = "high"
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
	os, err := hostInfo.LookupErr("extra", "versionString")
	if err != nil {
		return "", version.StringValue(), err
	}

	return os.StringValue(), version.StringValue(), nil
}
