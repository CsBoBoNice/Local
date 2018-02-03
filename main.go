package main

import (
	"fmt"
	"time"
	// CsDir "github.com/CsBoBoNice/Local/CsDir"
	CsSocket "github.com/CsBoBoNice/Local/CsSocket"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("nice day!")
	go CsSocket.ServerGo()
	time.Sleep(500 * time.Millisecond)
	go CsSocket.ClientGo(1)
	time.Sleep(10 * time.Second)

}

func ConnClose(chanServer chan string) {
	<-chanServer
	fmt.Printf("conn.Close()\n")
}

func scanfExit(chanServer chan string) {
	fmt.Printf("请输入:\n")
	var s1 string
	for {
		fmt.Scanln(&s1)
		switch s1 {
		case "exit":
			fmt.Printf("nice day~\n")
			chanServer <- "exit"
			return
		default:
			fmt.Printf("格式不对清重新输入\n")
		}
	}
}
