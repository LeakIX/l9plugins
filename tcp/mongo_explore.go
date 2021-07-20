package tcp

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

func (plugin MongoSchemaPlugin) Run(ctx context.Context, event *l9format.L9Event, pluginOptions map[string]string) (hasLeak bool) {
	log.Printf("Trying mongodb://%s", net.JoinHostPort(event.Ip,event.Port))
	mongoUrl := fmt.Sprintf("mongodb://%s/", net.JoinHostPort(event.Ip, event.Port))
	if event.HasTransport("tls") {
		mongoUrl += "?tls=true&tlsAllowInvalidCertificates=true&tlsInsecure=true"
	}
	client, err := mongo.Connect(
		ctx, options.Client().ApplyURI(mongoUrl).SetDialer(plugin))
	if err != nil {
		log.Println("Connect error: " + err.Error())
		return  false
	}
	defer client.Disconnect(nil)
	dbList, err := client.ListDatabases(ctx, bson.D{{}})
	if err != nil {
		log.Println("ListDB error: " + err.Error())
		return  false
	}
	for _, dbInfo := range dbList.Databases {
		db := client.Database(dbInfo.Name)
		collections, err := db.ListCollectionNames(ctx, bson.D{{}})
		if err != nil {
			continue
		}
		if strings.Contains(dbInfo.Name, "READ_ME") || strings.Contains(dbInfo.Name, "WARNING") || strings.Contains(dbInfo.Name, "meow") {
			event.Leak.Dataset.Infected = true
		}
		for _, collectionName := range collections {
			if strings.Contains(collectionName, "READ_ME") || strings.Contains(collectionName, "WARNING") || strings.Contains(collectionName, "meow") {
				event.Leak.Dataset.Infected = true
			}
			event.Leak.Dataset.Collections++
			event.Summary += fmt.Sprintf("Found collection %s.%s ", dbInfo.Name, collectionName)
			result := db.RunCommand(ctx, bson.D{{ "collStats" , collectionName}, {"scale", 1} })
			if result.Err() == nil {
				collectionStats := &MongoCollectionDetails{}
				err = result.Decode(&collectionStats)
				if err == nil {
					event.Leak.Dataset.Rows += collectionStats.Count
					event.Leak.Dataset.Size += collectionStats.Size
					event.Summary += fmt.Sprintf(" with %d documents (%s)", collectionStats.Count, utils.HumanByteCount(collectionStats.Size))
				} else {
					log.Println(err.Error())
				}
			} else {
				log.Println(result.Err().Error())
			}
			event.Summary += "\n"
			log.Printf("Found collection %s.%s\n", dbInfo.Name, collectionName)
		}
	}
	if event.Leak.Dataset.Collections < 1 {
		return false
	}
	event.Summary = fmt.Sprintf("Collections: %d, document count: %d, size: %s\n",
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

type MongoCollectionDetails struct {
	Count int64 `json:"count"`
	Size int64 `json:"storageSize"`
}