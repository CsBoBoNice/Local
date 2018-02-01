package main

import (
	// "os"
	// "bytes"
	"fmt"
	CsDir "github.com/CsBoBoNice/Local/CsDir"
	// CsSocket "github.com/CsBoBoNice/Local/CsSocket"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println("nice day!")
	// SrcDir := "E:/go1.jpg"
	// ok, err := CsDir.PathExists(SrcDir)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// if ok {
	// 	fmt.Println("Path Exists!")
	// } else {
	// 	fmt.Println("Path not exist!")
	// }

	SrcDir, BuildDir, Suffix := CsDir.DirInitLocal()

	var s_walkdir CsDir.Walkdir_s
	s_walkdir.WalkDirFile(SrcDir, BuildDir, Suffix)
	fmt.Println(s_walkdir.DirHead)
	// fmt.Println(s_walkdir.FileMD5)

	// for i, v := range s_walkdir.FileMD5 {
	// 	md5, name := CsDir.UnpackFileMD5(v)
	// 	fmt.Println(i, "\n", md5, "\n", name, "\n", len(v), "\n")
	// }
	// fmt.Println(len(s_walkdir.FileMD5))

	byteDate := CsDir.PackSliceString(s_walkdir.Files)
	// fmt.Printf("%s", byteDate)
	// fmt.Println(len(byteDate))
	stringDate := CsDir.UnpackSliceString(byteDate)
	// fmt.Println(stringDate)
	fmt.Println(stringDate[1] == s_walkdir.Files[1])

	// CsDir.ContrastDir(local, Backup)

	// // reader := bufio.NewReader()
	// reader := bufio.NewScanner()
	// line, err := reader.ReaderBytes()

	// var buffer bytes.Buffer
	// var buff []byte
	// for i := 66; i < 100; i++ {
	// 	buffer.WriteByte(byte(i))
	// }

	// buff = buffer.Bytes()
	// fmt.Println(buff)

	// var outDate_s CsSocket.Data
	// var inDate_s CsSocket.Data
	// outDate_s.DataBuff = buff
	// outDate_s.PackData()
	// inDate_s.DataHeadbuff = outDate_s.DataHeadbuff
	// fmt.Println("outDate_s", outDate_s.DataHeadbuff)
	// fmt.Println("inDate_s", inDate_s.DataHeadbuff)
	// inDate_s.UnpackData(inDate_s.DataHeadbuff)
	// fmt.Println("inDate_s\t", inDate_s.Datahead.DataSize, "\t", inDate_s.Datahead.MD5Byte)

	// chanServer := make(chan string)

	// go scanfExit(chanServer)

	// ConnClose(chanServer)
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

	// fmt.Printf("nice day~\n")
}
