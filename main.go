package main

import (
	"bytes"
	"fmt"
	CsSocket "github.com/CsBoBoNice/Local/CsSocket"
)

func main() {
	fmt.Println("nice day!")
	// // reader := bufio.NewReader()
	// reader := bufio.NewScanner()
	// line, err := reader.ReaderBytes()
	var buffer bytes.Buffer
	var buff []byte
	for i := 66; i < 100; i++ {
		buffer.WriteByte(byte(i))
	}

	buff = buffer.Bytes()
	fmt.Println(buff)

	var outDate_s CsSocket.Data
	var inDate_s CsSocket.Data
	outDate_s.DataBuff = buff
	outDate_s.PackData()
	inDate_s.DataHeadbuff = outDate_s.DataHeadbuff
	fmt.Println("outDate_s", outDate_s.DataHeadbuff)
	fmt.Println("inDate_s", inDate_s.DataHeadbuff)
	inDate_s.UnpackData(inDate_s.DataHeadbuff)
	fmt.Println("inDate_s\t", inDate_s.Datahead.DataSize, "\t", inDate_s.Datahead.MD5Byte)

}
