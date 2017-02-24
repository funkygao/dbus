package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/funkygao/golib/color"
	"github.com/funkygao/golib/io"
)

const width = 17

// Usage:
// go run kpub.go -z test -c test -t test -ack all -n 1024 -sz 25
// gk peek -z test -c test -t test -body > test.txt
// go run validator.go test.txt
func main() {
	if len(os.Args) != 2 {
		panic("Usage go run validator.go <filename>")
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}

	min, max := 1<<10, 0
	lineN := 0
	buf := make([]int, width)
	bufIdx := 0
	reader := bufio.NewReader(f)
	for {
		l, err := io.ReadLine(reader)
		if err != nil {
			// EOF
			break
		}

		// {000000019} XXXXXXXXXXXXXXXXXXXXXXXXX
		n, err := strconv.Atoi(string(l[1:10]))
		if err != nil {
			println(string(l), err.Error())
			break
		}

		if n > max {
			max = n
		}
		if n < min {
			min = n
		}

		buf[bufIdx] = n
		bufIdx = (bufIdx + 1) % width
		lineN++
		if lineN%width == 0 {
			// output the buf
			last := buf[0]
			s := fmt.Sprintf("%9d ", last)
			for _, d := range buf[1:] {
				if d == last {
					s += color.Red("%9d ", d)
				} else {
					s += fmt.Sprintf("%9d ", d)
				}
			}
			fmt.Println(s)

			buf = buf[0:]
		}

	}

	fmt.Printf("min=%d, max=%d\n", min, max)
}
