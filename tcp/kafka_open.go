package tcp

import (
	"context"
	"fmt"
	"github.com/LeakIX/l9format"
	"github.com/Shopify/sarama"
	"log"
	"net"
	"time"
)

type KafkaOpenPlugin struct {
	l9format.ServicePluginBase
}

func (KafkaOpenPlugin) GetVersion() (int, int, int) {
	return 0, 0, 1
}

func (KafkaOpenPlugin) GetProtocols() []string {
	return []string{"kafka"}
}

func (KafkaOpenPlugin) GetName() string {
	return "KafkaOpenPlugin"
}

func (KafkaOpenPlugin) GetStage() string {
	return "open"
}

// Get info
func (plugin KafkaOpenPlugin) Run(ctx context.Context, event *l9format.L9Event, pluginOptions map[string]string) (hasLeak bool) {
	config := sarama.NewConfig()
	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		config.Net.DialTimeout = deadline.Sub(time.Now())
	} else {
		config.Net.DialTimeout = 5 * time.Second
	}
	config.Consumer.Return.Errors = true

	//kafka end point
	brokers := []string{net.JoinHostPort(event.Ip, event.Port)}
	cluster, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		log.Println(err)
		return false
	}
	defer cluster.Close()
	topics, err := cluster.Topics()
	if err != nil || len(topics) < 1 {
		log.Println(err)
		return false
	}
	event.Summary = "NoAuth\n"
	for _, topic := range topics {
		event.Summary += fmt.Sprintf("Found topic %s\n", topic)
	}
	return true
}
