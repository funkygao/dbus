// A smoke test for batcher package.
package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/funkygao/dbus/pkg/batcher"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/golib/color"
	"github.com/funkygao/golib/gofmt"
	"github.com/funkygao/golib/sampling"
	"github.com/funkygao/golib/sequence"
)

var (
	benchmark bool
	msgN      = 1 << 5
	batchSize = 5
)

type producer struct {
	wg sync.WaitGroup

	batcher *batcher.Batcher

	kafkaOutDatas []int
	kafkaOut      chan int

	kafkaSentOk *sequence.Sequence

	failBatchN int
	okN        int
}

func newProducer() *producer {
	p := &producer{
		batcher:       batcher.NewBatcher(batchSize),
		kafkaOutDatas: []int{},
		kafkaOut:      make(chan int, 25),
		kafkaSentOk:   sequence.New(),
	}

	p.wg.Add(1)
	go p.batcherWorker()

	p.wg.Add(1)
	go p.kafkaWorker()

	return p
}

func (p *producer) Send(i int) {
	p.batcher.Write(i)
}

func (p *producer) batcherWorker() {
	defer p.wg.Done()

	history := make(map[int]struct{})
	for {
		msg, err := p.batcher.ReadOne()
		if err == batcher.ErrStopping {
			close(p.kafkaOut)
			break
		}

		i := msg.(int)

		if !benchmark {
			if _, present := history[i]; present {
				log.Println(color.Yellow("batcher read <- %d", i))
			} else {
				log.Printf("batcher read <- %d", i)
			}
			history[i] = struct{}{}

			// got data from ringbatcherfer and pass it to kafka worker
			// if some batch fails, ringbatcherfer will feed us again
			p.kafkaOutDatas = append(p.kafkaOutDatas, i)
		}

		p.kafkaOut <- i
	}
}

func (p *producer) kafkaWorker() {
	defer p.wg.Done()

	batchIdx := 0
	batch := make([]int, batchSize)
	for i := range p.kafkaOut {
		batch[batchIdx] = i

		batchIdx++
		if batchIdx == batchSize {
			if !benchmark && sampling.SampleRateSatisfied(1800) {
				// fails
				p.failBatchN++
				log.Println(color.Red("%+v", batch))

				for j := 0; j < batchSize; j++ {
					p.batcher.Fail()
				}
				log.Println(color.Red("%+v", p.batcher))
			} else {
				// success
				log.Println(color.Green("%+v", batch))

				for j := 0; j < batchSize; j++ {
					p.batcher.Succeed()
					p.okN++

					if !benchmark {
						p.kafkaSentOk.Add(batch[j])
					}
				}
			}

			batchIdx = 0
			batch = batch[0:]
		}
	}
}

func (p *producer) Close() {
	p.wg.Wait()
}

func (p *producer) PrintResult() {
	log.Printf("failed=%d, msg=%d, kafka sent ok=%d, kafka tried=%d", p.failBatchN, msgN, p.kafkaSentOk.Length(), len(p.kafkaOutDatas))
	min, max, loss := p.kafkaSentOk.Summary()
	log.Println("kafka sent ok:", p.kafkaSentOk, min, max)
	log.Println("kafka tried:", p.kafkaOutDatas)

	if len(loss) == 0 {
		log.Println(color.Green("passed after %d retries", p.failBatchN))
	} else {
		log.Println(color.Red("lost %d", len(loss)))
		log.Println(loss)
	}

}

func init() {
	flag.BoolVar(&benchmark, "b", false, "benchmark mode")
	flag.Parse()
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if benchmark {
		if true {
			log.SetOutput(ioutil.Discard)

		}
		msgN = 10 << 20
		batchSize = 2 << 10
	} else {
		log.SetOutput(os.Stdout)
	}

	agent.Start()

	inChan := make(chan int)

	p := newProducer()

	t0 := time.Now()
	go func() {
		for i := 0; i < msgN; i++ {
			inChan <- i
		}

		close(inChan)
		p.batcher.Close()
	}()

	for i := range inChan {
		p.Send(i)
	}

	p.Close()
	if benchmark {
		println(gofmt.Comma(int64(p.okN)/int64(time.Since(t0).Seconds())), "per second")
	} else {
		p.PrintResult()
	}
}
