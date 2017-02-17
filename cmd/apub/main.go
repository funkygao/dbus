// A script to test kafka async and ack mechanism.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/dbus/pkg/kafka"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/color"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/signal"
	"github.com/funkygao/golib/sync2"
	"github.com/funkygao/log4go"
)

var (
	zone, cluster, topic string
	ack                  string
	syncMode             bool
	maxErrs              int64
	messages             int
)

func init() {
	ctx.LoadFromHome()

	flag.StringVar(&zone, "z", "prod", "zone")
	flag.StringVar(&cluster, "c", "", "cluster")
	flag.StringVar(&topic, "t", "", "topic")
	flag.StringVar(&ack, "ack", "local", "local|none|all")
	flag.BoolVar(&syncMode, "sync", false, "sync mode")
	flag.Int64Var(&maxErrs, "e", 10, "max errors before quit")
	flag.IntVar(&messages, "n", 2000, "flush messages")
	flag.Parse()

	if len(zone) == 0 || len(cluster) == 0 || len(topic) == 0 {
		panic("invalid flag")
	}

	sarama.Logger = log.New(os.Stdout, color.Magenta("[Sarama]"), log.LstdFlags|log.Lshortfile)
	log4go.SetLevel(log4go.TRACE)
}

func main() {
	cf := kafka.DefaultConfig()
	cf.Sarama.Producer.Flush.Messages = messages
	if syncMode {
		cf.SyncMode()
	}
	switch ack {
	case "none":
		cf.Ack(sarama.NoResponse)
	case "local":
		cf.Ack(sarama.WaitForLocal)
	case "all":
		cf.Ack(sarama.WaitForAll)
	default:
		panic("invalid: " + ack)
	}
	p, err := kafka.NewProducer("tester", zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone))).NewCluster(cluster).BrokerList(), cf)
	if err != nil {
		panic(err)
	}

	var (
		sent, sentOk sync2.AtomicInt64
	)

	p.SetErrorHandler(func(err *sarama.ProducerError) {
		v, _ := err.Msg.Value.Encode()
		log.Println(string(v[:12]), err)
	})
	p.SetSuccessHandler(func(msg *sarama.ProducerMessage) {
		sentOk.Add(1)
	})
	if err := p.Start(); err != nil {
		panic(err)
	}

	closed := make(chan struct{})
	var once sync.Once
	signal.RegisterHandler(func(sig os.Signal) {
		log.Printf("got signal %s", sig)

		once.Do(func() {
			close(closed)
		})
	}, syscall.SIGINT)

	go func() {
		for {
			log.Println(gofmt.Comma(sent.Get()), gofmt.Comma(sentOk.Get()), gofmt.Comma(sent.Get()-sentOk.Get()))
			time.Sleep(time.Second)
		}
	}()

	for {
		select {
		case <-closed:
			goto BYE
		default:
		}

		msg := sarama.StringEncoder(fmt.Sprintf("{%d} %s", sent.Get(),
			strings.Repeat("X", 10331)))
		if err := p.Send(&sarama.ProducerMessage{
			Topic: topic,
			Value: msg,
		}); err != nil {
			log.Println(err)
			break
		}

		sent.Add(1)
	}

BYE:
	log.Printf("%d/%d, closed with %v", sentOk.Get(), sent.Get(), p.Close())

}
