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
	"github.com/funkygao/gafka/ctx"
	"github.com/funkygao/gafka/zk"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/signal"
	"github.com/funkygao/golib/sync2"
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

	log.SetFlags(log.LstdFlags | log.Lshortfile)
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
	//cf.Producer.Flush.Bytes = 1 << 20
	//sarama.MaxRequestSize = 1 << 20
	//cf.Producer.MaxMessageBytes = 1 << 20
	cf.Producer.Flush.MaxMessages = 0 // unlimited

	zkzone := zk.NewZkZone(zk.DefaultConfig(zone, ctx.ZoneZkAddrs(zone)))
	zkcluster := zkzone.NewCluster(cluster)
	p, err := sarama.NewAsyncProducer(zkcluster.BrokerList(), cf)
	if err != nil {
		panic(err)
	}

	var (
		sent, sentOk sync2.AtomicInt64
	)

	stopProducer := make(chan struct{})
	closed := make(chan struct{})
	var once sync.Once
	quitFn := func() {
		once.Do(func() {
			close(stopProducer)
			time.Sleep(time.Second)

			log.Printf("%d/%d, closed with %v\n", sentOk.Get(), sent.Get(), p.Close())
			close(closed)
		})
	}

	signal.RegisterHandler(func(sig os.Signal) {
		log.Printf("got signal %s", sig)
		quitFn()
	}, syscall.SIGINT)

	go func() {
		var errN int64
		for {
			select {
			case _, ok := <-p.Successes():
				if !ok {
					// p.Close will eat up the success messages
					log.Println("end of success")
					return
				}

				sentOk.Add(1)

			case err, ok := <-p.Errors():
				if !ok {
					log.Println("end of error")
					return
				}

				v, _ := err.Msg.Value.Encode()
				log.Println(string(v[:15]), err)

				errN++
				if maxErrs > 0 && errN >= maxErrs {
					quitFn()
				}
			}
		}
	}()

	go func() {
		for {
			log.Println(gofmt.Comma(sent.Get()), gofmt.Comma(sentOk.Get()), gofmt.Comma(sent.Get()-sentOk.Get()))
			time.Sleep(time.Second)
		}
	}()

	for {
		msg := sarama.StringEncoder(fmt.Sprintf("{%d} %s", sent.Get(), strings.Repeat("X", 10331)))
		select {
		case <-stopProducer:
			goto BYE

		case p.Input() <- &sarama.ProducerMessage{
			Topic: topic,
			Value: msg,
		}:
			sent.Add(1)
		}
	}

BYE:
	<-closed
}
