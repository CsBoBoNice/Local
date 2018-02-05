package main

import (
	"fmt"
	CsDir "github.com/CsBoBoNice/Local/CsDir"
	CsSocket "github.com/CsBoBoNice/Local/CsSocket"
	"runtime"
	"time"
)

const (
	SERVER_NETWORK = "tcp"
	SERVER_ADDRESS = "192.168.31.67:8085"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	chanServer := make(chan string)

	fmt.Println("nice day!")
	go CsSocket.ServerGo(SERVER_NETWORK, SERVER_ADDRESS)
	time.Sleep(500 * time.Millisecond)

	//初始化本地读取文件夹，服务器需要备份的文件夹，还有要查找的文件后缀
	SrcDir, BackupDir, _ := CsDir.DirInitRemote()
	go CsSocket.ClientGo(1, SERVER_NETWORK, SERVER_ADDRESS, SrcDir, BackupDir)

	go scanfExit(chanServer)
	<-chanServer
	fmt.Println("任务完成")
	// CsDir.DeleteDir(CsDir.JointDir2(SrcDir, BackupDir))	//测试用，将传过来的文件删除

	// time.Sleep(10 * time.Second)

}

func ConnClose(chanServer chan string) {
	<-chanServer
	fmt.Printf("conn.Close()\n")
}

func scanfExit(chanServer chan string) {
	// fmt.Printf("请输入:\n")
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
