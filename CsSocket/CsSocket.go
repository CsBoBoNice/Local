package CsSocket

import (
	_ "bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	CsDir "github.com/CsBoBoNice/Local/CsDir"
	"net"
	"strings"
	"time"
)

const (
	SERVER_NETWORK = "tcp"
	SERVER_ADDRESS = "127.0.0.1:8085"
)

type Data struct {
	Datahead     DataHead
	DataHeadbuff []byte
	DataBuff     []byte
}

type DataHead struct {
	DataSize uint64
	MD5Byte  [16]byte
}

func (Datas *Data) PackData() {
	Datas.Datahead.DataSize = uint64(len(Datas.DataBuff))
	Datas.Datahead.MD5Byte = md5.Sum(Datas.DataBuff)
	DataSize := Uint64ToByte(Datas.Datahead.DataSize)
	for _, r := range DataSize {
		Datas.DataHeadbuff = append(Datas.DataHeadbuff, r)
	}
	for _, r := range Datas.Datahead.MD5Byte {
		Datas.DataHeadbuff = append(Datas.DataHeadbuff, r)
	}
	// fmt.Println(Datas.DataHeadbuff)
}

func (Datas *Data) UnpackData(date []byte) {
	Datas.DataHeadbuff = date
	Datas.Datahead.DataSize = ByteToUint64(Datas.DataHeadbuff[0:8])
	for i, d := range Datas.DataHeadbuff[8:] {
		Datas.Datahead.MD5Byte[i] = d
	}
}

func ByteToUint64(date []byte) (i uint64) {
	i = uint64(binary.BigEndian.Uint64(date[0:8]))
	// fmt.Println(i)
	return
}

func Uint64ToByte(i uint64) (date []byte) {
	date = make([]byte, 8)
	binary.BigEndian.PutUint64(date, uint64(i))
	return
}

func printLog(role string, sn int, format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Printf("%s[%d]: %s", role, sn, fmt.Sprintf(format, args...))
}

func printServerLog(format string, args ...interface{}) {
	printLog("Server", 0, format, args...)
}

func printClientLog(sn int, format string, args ...interface{}) {
	printLog("Client", sn, format, args...)
}

func readHead(conn net.Conn) ([]byte, error) {
	return read(conn, 24)
}

func read(conn net.Conn, num uint64) ([]byte, error) {
	readBytes := make([]byte, 1)
	var buffer bytes.Buffer
	var readSize uint64 = 0
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return buffer.Bytes(), err
		} else {
			readSize++
		}
		readByte := readBytes[0]
		buffer.WriteByte(readByte)
		if readSize >= num {
			break
		}
	}
	return buffer.Bytes(), nil
}

func write(conn net.Conn, date []byte) (int, error) {
	var buffer bytes.Buffer
	buffer.Write(date)
	return conn.Write(buffer.Bytes())
}

//按照协议读取
func ReadAgreement(conn net.Conn) (buff []byte, err error) {
	var date Data
	Headbuff, err := readHead(conn) //读取传输数据头
	if err != nil {
		printServerLog("Accept Error: %s", err)
	}

	date.UnpackData(Headbuff) //解压传输数据头

	date.DataBuff, err = read(conn, date.Datahead.DataSize) //读取真实数据
	if err != nil {
		printServerLog("Accept Error: %s", err)
	}

	MD5Byte := md5.Sum(date.DataBuff)
	if MD5Byte != date.Datahead.MD5Byte {
		printServerLog("Accept Error: %s", "发送与接收数据不符")
		return
	}
	buff = date.DataBuff
	return
}

//按照协议写入
func WriteAgreement(conn net.Conn, buff []byte) (err error) {
	var date Data
	date.DataBuff = buff
	date.PackData()
	_, err = conn.Write(date.DataHeadbuff)
	if err != nil {
		printServerLog("Accept Error: %s", err)
	}
	_, err = conn.Write(date.DataBuff)
	if err != nil {
		printServerLog("Accept Error: %s", err)
	}
	return
}

func ServerGo() {
	var listener net.Listener
	listener, err := net.Listen(SERVER_NETWORK, SERVER_ADDRESS)
	if err != nil {
		printServerLog("Listen Error: %s", err)
		return
	}
	defer listener.Close()
	printServerLog("Got listener for the server. (local address: %s)", listener.Addr())
	for {
		conn, err := listener.Accept() // 阻塞直至新连接到来。
		if err != nil {
			printServerLog("Accept Error: %s", err)
		}
		printServerLog("Established a connection with a client application. (remote address: %s)",
			conn.RemoteAddr())
		go handleConn(conn)
	}
}

//服务端有连接处理代码
func handleConn(conn net.Conn) {
	for {
		conn.SetDeadline(time.Now().Add(15 * time.Second))
		SrcDir, BuildDir, Suffix := CsDir.DirInitLocal() //初始化本地读取文件夹，远端创建的文件夹，还有要查找的文件后缀
		
		var local CsDir.Walkdir_s
		local.WalkDirFile(SrcDir, BuildDir, Suffix)                  //遍历本地目录
		WriteAgreement(conn, CsDir.PackSliceString(local.TargetDir)) //将本地的所有目标目录发给远端

		WriteAgreement(conn, CsDir.PackSliceString(local.FileMD5)) //将本地的 包含MD5码的文件目录 发给远端

		dir, _ := ReadAgreement(conn)       //接收远端的所有目标文件目录
		Dir := CsDir.UnpackSliceString(dir) //解析出所有目标文件目录FileMD5

		// WriteAgreement(conn, []byte("Start the transfer file!")) //将开始标志发给远端
		for _, v := range Dir {
			WriteAgreement(conn, []byte(v))                                           //将文件目录发给远端
			WriteAgreement(conn, CsDir.ReadFileAll(CsDir.JointDir(local.DirHead, v))) //将文件数据发给远端
		}
		WriteAgreement(conn, []byte("The transfer file is finished!")) //将结束标志发给远端
		return
	}
}

func ClientGo(id int) {
	//向指定的网络地址发送链接建立申请，并堵塞一段时间，超时则err!=nil
	conn, err := net.DialTimeout(SERVER_NETWORK, SERVER_ADDRESS, 2*time.Second)
	if err != nil {
		printClientLog(id, "Dial Error: %s", err)
		return
	}
	defer conn.Close()
	printClientLog(id, "Connected to server. (remote address: %s, local address: %s)",
		conn.RemoteAddr(), conn.LocalAddr())

	//初始化本地读取文件夹，远端创建的文件夹，还有要查找的文件后缀
	SrcDir, BuildDir, Suffix := CsDir.DirInitRemote()

	var local CsDir.Walkdir_s
	var Backup CsDir.Walkdir_s
	var LocalNow CsDir.Walkdir_s

	targetDir, err := ReadAgreement(conn)                 //接收远端的所有目标目录
	Backup.TargetDir = CsDir.UnpackSliceString(targetDir) //解析出所有目标目录

	SrcDirNow := CsDir.JointDir(SrcDir, Backup.TargetDir[0])
	local.WalkDirFile(SrcDirNow, BuildDir, Suffix) //遍历本地目录

	//对比本地目录与远端目录，以发送过来的远端目录为基准，将多余的，目录删除，不足的目录新建
	CsDir.ContrastDir(local.TargetDir, Backup.TargetDir, local.DirHead)

	LocalNow.WalkDirFile(SrcDirNow, BuildDir, Suffix) //遍历本地目录

	targetFile, err := ReadAgreement(conn)               //接收远端的所有目标文件目录
	Backup.FileMD5 = CsDir.UnpackSliceString(targetFile) //解析出所有目标文件目录FileMD5

	//解析出包含MD5码的文件目录，格式为MD5+TargetFile
	//将没有匹配文件的 与 MD5码与文件不同的目录找出
	Dir := CsDir.ContrastDirMD5(LocalNow.FileMD5, Backup.FileMD5, LocalNow.DirHead)

	WriteAgreement(conn, CsDir.PackSliceString(Dir)) //将没有匹配文件的 与 MD5码与文件不同的目录找出发给远端
	var dirName string
	for {
		dast, _ := ReadAgreement(conn) //接收数据
		if string(dast) == "The transfer file is finished!" {
			break
		}

		dirName = CsDir.JointDir(SrcDir, string(dast))
		dast, _ = ReadAgreement(conn) //接收数据
		CsDir.WriteFileAll(dirName, dast)
	}

	printClientLog(id, "Client close. (remote address: %s)", conn.LocalAddr())

}
