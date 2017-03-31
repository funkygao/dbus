package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/funkygao/golib/color"
	"github.com/funkygao/golib/io"
)

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
	reader := bufio.NewReader(f)
	last := -98734
	for {
		l, err := io.ReadLine(reader)
		if err != nil {
			// EOF
			break
		}

		// {000000019} XXXXXXXXXXXXXXXXXXXXXXXXX
		n, err := strconv.Atoi(string(l[1:10]))
		if err != nil {
			fmt.Printf("%s for `%s`\n", err, string(l))
			break
		}

		if n > max {
			max = n
		}
		if n < min {
			min = n
		}
		lineN++

		if last >= 0 && n != last+1 {
			fmt.Println(color.Red("%d %d %d", last, n, n-last))
		}

		last = n
	}

	fmt.Printf("min=%d, max=%d, lines=%d\n", min, max, lineN)
	if lineN-max == 1 {
		// kpub always starts with 0
		fmt.Println(color.Green("passed"))
	} else {
		fmt.Println(color.Red("failed"))
	}
}
