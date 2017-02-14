// A script to test kafka async and ack mechanism.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/signal"
)

var (
	zone, cluster, topic string
	ack                  string
	maxErrs              int64
	messages             int
)

func init() {
	ctx.LoadFromHome()

	flag.StringVar(&zone, "z", "prod", "zone")
	flag.StringVar(&cluster, "c", "", "cluster")
	flag.StringVar(&topic, "t", "", "topic")
	flag.StringVar(&ack, "ack", "local", "local|none|all")
	flag.Int64Var(&maxErrs, "e", 10, "max errors before quit")
	flag.IntVar(&messages, "n", 2000, "flush messages")
	flag.Parse()

	if len(zone) == 0 || len(cluster) == 0 || len(topic) == 0 {
		panic("invalid flag")
	}
}

func main() {
	cf := sarama.NewConfig()
	cf.ChannelBufferSize = 256 * 4 // default was 256
	cf.Producer.Return.Errors = true
	cf.Producer.Return.Successes = true
	cf.Producer.Retry.Backoff = time.Millisecond * 300
	cf.Producer.Retry.Max = 3
	switch ack {
	case "none":
		cf.Producer.RequiredAcks = sarama.NoResponse
	case "local":
		cf.Producer.RequiredAcks = sarama.WaitForLocal
	case "all":
		cf.Producer.RequiredAcks = sarama.WaitForAll
	default:
		panic("invalid: " + ack)
	}
	cf.Producer.RequiredAcks = sarama.WaitForLocal
	cf.Producer.Flush.Frequency = time.Second
	cf.Producer.Flush.Messages = messages
	cf.Producer.Flush.MaxMessages = 0 // unlimited

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	zkcluster := zkzone.NewCluster(cluster)
	p, err := sarama.NewAsyncProducer(zkcluster.BrokerList(), cf)
	if err != nil {
		panic(err)
	}

	quit := make(chan struct{})
	closed := make(chan struct{})
	signal.RegisterHandler(func(sig os.Signal) {
		close(quit)
		time.Sleep(time.Second)

		fmt.Printf("closed with %v\n", p.Close())
		close(closed)
	}, syscall.SIGINT)

	var (
		sent, sentOk int64
	)

	go func() {
		var errN int64
		for {
			select {
			case <-quit:
				return

			case <-p.Successes():
				sentOk++

			case err := <-p.Errors():
				v, _ := err.Msg.Value.Encode()
				fmt.Println(string(v[:15]), err)

				errN++
				if maxErrs > 0 && errN >= maxErrs {
					close(quit)
					time.Sleep(time.Second)

					fmt.Printf("closed with %v\n", p.Close())
					close(closed)
				}
			}
		}
	}()

	go func() {
		for {
			fmt.Println(gofmt.Comma(sent), gofmt.Comma(sentOk), gofmt.Comma(sent-sentOk))
			time.Sleep(time.Second)
		}
	}()

	for {
		msg := sarama.StringEncoder(fmt.Sprintf("{%d} %s", sent, strings.Repeat("X", 1033)))
		select {
		case <-quit:
			goto BYE

		case p.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: msg,
		}:
			sent++
		}
	}

BYE:
	<-closed
}
