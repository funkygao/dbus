// A smoke test for batcher package.
package main

import (
	"log"
	"os"
	"sync"

	"github.com/funkygao/dbus/pkg/batcher"
	"github.com/funkygao/gafka/diagnostics/agent"
	"github.com/funkygao/golib/color"
	"github.com/funkygao/golib/sampling"
	"github.com/funkygao/golib/sequence"
)

const (
	msgN        = 1 << 5
	batcherSize = 16
	batchSize   = 5
)

type producer struct {
	wg sync.WaitGroup

	batcher *batcher.Batcher

	kafkaOutDatas []int
	kafkaOut      chan int

	kafkaSentOk *sequence.Sequence

	failBatchN int
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
		if _, present := history[i]; present {
			log.Println(color.Yellow("batcher read <- %d", i))
		} else {
			log.Printf("batcher read <- %d", i)
		}
		history[i] = struct{}{}

		// got data from ringbatcherfer and pass it to kafka worker
		// if some batch fails, ringbatcherfer will feed us again
		p.kafkaOutDatas = append(p.kafkaOutDatas, i)
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
			if sampling.SampleRateSatisfied(1800) {
				// fails
				p.failBatchN++
				log.Println(color.Red("%+v", batch))

				for j := 0; j < batchSize; j++ {
					p.batcher.Rollback()
				}
				log.Println(color.Red("%+v", p.batcher))
			} else {
				// success
				log.Println(color.Green("%+v", batch))

				for j := 0; j < batchSize; j++ {
					p.batcher.Advance()
					p.kafkaSentOk.Add(batch[j])
				}
				log.Println(color.Green("%+v", p.batcher))
			}

			batchIdx = 0
			batch = batch[0:]
		}
	}
}

func (p *producer) close() {
	p.wg.Wait()
}

func (p *producer) printResult() {
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

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	agent.Start()

	inChan := make(chan int)

	p := newProducer()

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

	p.close()
	p.printResult()
}
