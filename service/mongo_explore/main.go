package main

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/LeakIX/l9format/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"strings"
)

type MongoSchemaPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return MongoSchemaPlugin{}
}

func (MongoSchemaPlugin) GetVersion() (int, int, int) {
	return 0, 1, 1
}

func (MongoSchemaPlugin) GetProtocols() []string {
	return []string{"mongo"}
}

func (MongoSchemaPlugin) GetName() string {
	return "MongoSchemaPlugin"
}

func (MongoSchemaPlugin) GetStage() string {
	return "explore"
}

func (plugin MongoSchemaPlugin) Run(ctx context.Context, event *l9format.L9Event, pluginOptions map[string]string) (leak l9format.L9LeakEvent, hasLeak bool) {
	log.Printf("Trying mongodb://%s", net.JoinHostPort(event.Ip, event.Port))
	mongoUrl := fmt.Sprintf("mongodb://%s/", net.JoinHostPort(event.Ip, event.Port))
	if event.HasTransport("tls") {
		mongoUrl += "?tls=true&tlsAllowInvalidCertificates=true&tlsInsecure=true"
	}
	client, err := mongo.Connect(
		ctx, options.Client().ApplyURI(mongoUrl).SetDialer(plugin))
	if err != nil {
		log.Println("Connect error: " + err.Error())
		return leak, false
	}
	defer client.Disconnect(nil)
	dbList, err := client.ListDatabases(ctx, bson.D{{}})
	if err != nil {
		log.Println("ListDB error: " + err.Error())
		return leak, false
	}
	for _, dbInfo := range dbList.Databases {
		db := client.Database(dbInfo.Name)
		collections, err := db.ListCollectionNames(ctx, bson.D{{}})
		if err != nil {
			continue
		}
		if strings.Contains(dbInfo.Name, "READ_ME") || strings.Contains(dbInfo.Name, "WARNING") || strings.Contains(dbInfo.Name, "meow") {
			leak.Dataset.Infected = true
		}
		for _, collectionName := range collections {
			if strings.Contains(collectionName, "READ_ME") || strings.Contains(collectionName, "WARNING") || strings.Contains(collectionName, "meow") {
				leak.Dataset.Infected = true
			}
			leak.Dataset.Collections++
			leak.Data += "Found collection " + dbInfo.Name + "." + collectionName
			result := db.RunCommand(ctx, bson.D{{"collStats", collectionName}, {"scale", 1}})
			if result.Err() == nil {
				collectionStats := &MongoCollectionDetails{}
				err = result.Decode(&collectionStats)
				if err == nil {
					leak.Dataset.Rows += collectionStats.Count
					leak.Dataset.Size += collectionStats.Size
					leak.Data += fmt.Sprintf(" with %d documents (%s)", collectionStats.Count, utils.HumanByteCount(collectionStats.Size))
				} else {
					log.Println(err.Error())
				}
			} else {
				log.Println(result.Err().Error())
			}
			leak.Data += "\n"
			log.Println("Found collection " + dbInfo.Name + "." + collectionName)
		}
	}
	if leak.Dataset.Collections < 1 {
		return leak, false
	}
	leak.Data = fmt.Sprintf("Collections: %d, document count: %d, size: %s\n",
		leak.Dataset.Collections, leak.Dataset.Rows, utils.HumanByteCount(leak.Dataset.Size)) +
		leak.Data
	return leak, true
}

type MongoCollectionDetails struct {
	Count int64 `json:"count"`
	Size  int64 `json:"storageSize"`
}
