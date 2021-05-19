package tcp

import (
	"context"
	"github.com/LeakIX/l9format"
	"github.com/go-redis/redis/v8"
	"log"
	"net"
	"strings"
)

type RedisOpenPlugin struct {
	l9format.ServicePluginBase
}

func New() l9format.ServicePluginInterface {
	return RedisOpenPlugin{}
}

func (RedisOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (RedisOpenPlugin) GetProtocols() []string {
	return []string{"redis"}
}

func (RedisOpenPlugin) GetName() string {
	return "RedisOpenPlugin"
}

func (RedisOpenPlugin) GetStage() string {
	return "open"
}

func (plugin RedisOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, options map[string]string) bool {
	client := redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(event.Ip, event.Port),
		Password: "", // no password set
		DB:       0,  // use default DB
		Dialer:   plugin.DialContext,
	})
	defer client.Close()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Println("Redis PING failed, leaving early : ", err)
		return  false
	}
	redisInfo, err := client.Info(ctx).Result()
	if err != nil {
		log.Println("Redis INFO failed, leaving early : ", err)
		return false
	}
	redisInfoDict := make(map[string]string)
	redisInfo = strings.Replace(redisInfo, "\r", "", -1)
	for _, line := range strings.Split(redisInfo, "\n") {
		keyValuePair := strings.Split(line, ":")
		if len(keyValuePair) == 2 {
			redisInfoDict[keyValuePair[0]] = keyValuePair[1]
		}
	}
	if _, found := redisInfoDict["redis_version"]; found {
		event.Service.Software.OperatingSystem, _ = redisInfoDict["os"]
		event.Service.Software.Name = "Redis"
		event.Service.Software.Version, _ = redisInfoDict["redis_version"]
		event.Leak.Severity = l9format.SEVERITY_MEDIUM
		event.Summary = "Redis is open\n"
		event.Leak.Type = "open_database"
		event.Leak.Dataset.Rows = 1
		return true
	}
	return false
}
