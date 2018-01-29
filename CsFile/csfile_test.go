package Csfile

import (
	"fmt"
	"testing"
)

func TestCsfile(t *testing.T) {
	var err error
	const PartSize = 64
	//name := "E:/golang/GOPATH/src/filemaker/date.text"
	//name := "E:/golang/GOPATH/src/PartialFileProcessing/123.txt"
	//name := "E:/golang/GOPATH/src/PartialFileProcessing/main.go"
	name := "E:/golang/csfioletext/main.go"
	var Date []byte = make([]byte, PartSize)
	var file FilePartialProcessing
	file.InitProcessing(name, PartSize, 32) //初始化需要读的文件数据

	fmt.Println(file)

	err = file.ReadFileHead(Date) //读取文件头
	CheckFile(err)
	fmt.Printf("%s", Date)

	for i := 0; i < file.Fornum; i++ { //循环读取所有文件
		err = file.ReadPartFile(Date, i)
		CheckFile(err)
		fmt.Printf("%s", Date)

	}

	file.FileCloss()
	fmt.Println(file)
	fmt.Printf("size=%d\n", file.FileSize)
	fmt.Printf("OverDate=%d\n", file.OverDate)
	fmt.Printf("nice code\n")
	if err != nil {
		t.Error("bad code! try again!\n")
	}

}
