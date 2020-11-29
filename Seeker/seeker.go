package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/Knetic/govaluate"

	globals "github.com/prairir/JobProtocol/Globals"
)

func main() {
	// set timeout and connection
	conn, err := net.Dial(globals.CONN_TYPE, fmt.Sprint(globals.CONN_ADDR, ":", globals.CONN_PORT))
	fatalErrorCheck(err)

	// state
	// 0 initial connection
	// 1 waiting for HELLOACK
	// 2 first JOB EQN
	// 3 accepted and waiting for second JOB EQN
	// 4 closed
	state := 0
	// send HELLO at initial connection
	fmt.Fprintln(conn, "HELLO")
	fmt.Println("sent HELLO")
	state = 1
	for {
		fmt.Println("waiting for creator...")
		result, err := bufio.NewReader(conn).ReadString('\n')
		fmt.Println("received:", result)
		if err != nil {
			continue
		}
		cleanedResult := strings.TrimSpace(string(result))

		if state == 1 && cleanedResult == "HELLOACK" {
			state = 2
		} else if state == 2 && cleanedResult == "JOB EQN" {
			conn.Write([]byte("ACPT JOB EQN\r\n"))
			fmt.Println("received:", result)
			state = 3
		} else if state == 3 && cleanedResult[:7] == "JOB EQN" {
			fmt.Println("[", cleanedResult[8:], "]")
			expression, err := govaluate.NewEvaluableExpression(cleanedResult[8:])
			if err != nil {
				fmt.Println("job failed... bad input?", err.Error())
				conn.Write([]byte("JOB FAIL\r\n"))
				break
			}
			expResult, err := expression.Evaluate(nil)
			if err != nil {
				fmt.Println("job failed... bad input?", err.Error())
				conn.Write([]byte("JOB FAIL\r\n"))
				break
			} else {
				conn.Write([]byte("JOB SUCC " + fmt.Sprint(expResult) + "\r\n"))
			}
			state = 4
		}
		if state == 4 {
			state = 2
			continue
		}
	}
	conn.Close()
}

func fatalErrorCheck(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
