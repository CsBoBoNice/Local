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

type Data struct {
	Datahead     DataHead
	DataHeadbuff []byte
	DataBuff     []byte
}

type DataHead struct {
	DataSize uint64
	MD5Byte  [16]byte
}

var startTime time.Time

func InitTime() {
	startTime = time.Now()
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
	if len(date) <= 8 {
		fmt.Printf("error 123123123!")
	}
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
	fmt.Printf("%v\t%s[%d]: %s", time.Now().Sub(startTime), role, sn, fmt.Sprintf(format, args...))
	InitTime()
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

func ServerGo(network, address string) {
	var listener net.Listener
	listener, err := net.Listen(network, address)
	InitTime()
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
	defer func() {
		conn.Close()
	}()
	// conn.SetDeadline(time.Now().Add(15 * time.Second))
	// SrcDir, _ := CsDir.DirInitLocal() //初始化本地读取文件夹，远端创建的文件夹，还有要查找的文件后缀
	bytedir, _ := ReadAgreement(conn) //接收远端的所有目标文件目录
	SrcDir := string(bytedir)         //解读出本地需要遍历的目录
	ok, err := CsDir.IsDir(SrcDir)
	if err != nil {
		WriteAgreement(conn, []byte("The file can't be found!")) //本地找不到这个文件或目录
		printServerLog("本地找不到这个文件或目录")
		return
	}
	if ok != true {
		WriteAgreement(conn, []byte("Single file!")) //单一文件
		var com []byte
		var command string
		for {
			com, _ = ReadAgreement(conn) //接收远端的所有目标目录
			command = string(com)
			switch command {
			case "ok return!":
				return
			case "Give me MD5!":
				SingleMd5 := CsDir.GetMD5(SrcDir)
				md5V := SingleMd5[:]
				WriteAgreement(conn, md5V) //先将Md5码发过去
			case "Give me file!":
				WriteAgreement(conn, CsDir.ReadFileAll(SrcDir)) //将文件数据发给远端

			}
		}

	} else {
		WriteAgreement(conn, []byte("This is a Dir!")) //文件夹
	}
	var local CsDir.Walkdir_s
	local.WalkDirFile(SrcDir, "")                                //遍历本地目录
	WriteAgreement(conn, CsDir.PackSliceString(local.TargetDir)) //将本地的所有目标目录发给远端

	WriteAgreement(conn, CsDir.PackSliceString(local.FileMD5)) //将本地的 包含MD5码的文件目录 发给远端

	dir, _ := ReadAgreement(conn)       //接收远端的所有目标文件目录
	Dir := CsDir.UnpackSliceString(dir) //解析出所有目标文件目录FileMD5

	for _, v := range Dir {
		WriteAgreement(conn, []byte(v))                                           //将文件目录发给远端
		WriteAgreement(conn, CsDir.ReadFileAll(CsDir.JointDir(local.DirHead, v))) //将文件数据发给远端
	}
	WriteAgreement(conn, []byte("The transfer file is finished!")) //将结束标志发给远端
	return

}

func ClientGo(id int, network string, address string) {
	//向指定的网络地址发送链接建立申请，并堵塞一段时间，超时则err!=nil
	conn, err := net.DialTimeout(network, address, 2*time.Second)
	if err != nil {
		printClientLog(id, "Dial Error: %s", err)
		return
	}
	defer func() {
		printClientLog(id, "Client close. (remote address: %s)", conn.LocalAddr())
		conn.Close()
	}()

	printClientLog(id, "Connected to server. (remote address: %s, local address: %s)",
		conn.RemoteAddr(), conn.LocalAddr())

	//初始化本地读取文件夹，远端需要备份的文件夹，还有要查找的文件后缀
	SrcDir, BackupDir, _ := CsDir.DirInitRemote()

	WriteAgreement(conn, []byte(BackupDir)) //将服务器端需要备份的文件夹发过去

	com, _ := ReadAgreement(conn) //接收远端的所有目标目录
	command := string(com)
	if command == "The file can't be found!" {
		printClientLog(id, "服务器找不到这个文件)")
		return
	}

	if command == "Single file!" {
		targetDir := CsDir.GetTargetDir(BackupDir, CsDir.GetDirHead(BackupDir))
		SingleFile := CsDir.JointDir(SrcDir, targetDir)
		ok, _ := CsDir.PathExists(SingleFile)
		if ok {
			WriteAgreement(conn, []byte("Give me MD5!")) //让服务器将MD5码发过来
			SingleMd5get, _ := ReadAgreement(conn)       //接收MD5码
			singleMd5 := CsDir.GetMD5(SingleFile)        //读取本地Md5对比
			md5V := singleMd5[:]
			if string(md5V) == string(SingleMd5get) {
				WriteAgreement(conn, []byte("ok return!")) //文件相同可以退出了
				return
			} else {
				WriteAgreement(conn, []byte("Give me file!")) //让服务器将文件发过来
				buff, _ := ReadAgreement(conn)                //接收数据
				CsDir.WriteFileAll(SingleFile, buff)
			}
		} else {
			WriteAgreement(conn, []byte("Give me file!")) //让服务器将文件发过来
			buff, _ := ReadAgreement(conn)                //接收数据
			CsDir.WriteFileAll(SingleFile, buff)
		}
		WriteAgreement(conn, []byte("ok return!")) //文件相同可以退出了
		return
	}

	var local CsDir.Walkdir_s
	var Backup CsDir.Walkdir_s
	var LocalNow CsDir.Walkdir_s
	printClientLog(id, "正在接收远端的所有目标目录")
	targetDir, err := ReadAgreement(conn) //接收远端的所有目标目录
	printClientLog(id, "正在解析远端的所有目标目录")
	Backup.TargetDir = CsDir.UnpackSliceString(targetDir) //解析出所有目标目录

	if len(Backup.TargetDir) <= 0 {
		fmt.Println("没有目录！")
	}
	SrcDirNow := CsDir.JointDir(SrcDir, Backup.TargetDir[0])
	printClientLog(id, "正在遍历本地目录")
	local.WalkDirFile(SrcDirNow, "") //遍历本地目录

	//对比本地目录与远端目录，以发送过来的远端目录为基准，将多余的，目录删除，不足的目录新建
	CsDir.ContrastDir(local.TargetDir, Backup.TargetDir, local.DirHead)
	printClientLog(id, "正在遍历本地目录")
	LocalNow.WalkDirFile(SrcDirNow, "") //遍历本地目录
	printClientLog(id, "正在接收远端的所有目标文件目录")
	targetFile, err := ReadAgreement(conn) //接收远端的所有目标文件目录
	printClientLog(id, "正在解析出所有目标文件目录FileMD5")
	Backup.FileMD5 = CsDir.UnpackSliceString(targetFile) //解析出所有目标文件目录FileMD5

	//解析出包含MD5码的文件目录，格式为MD5+TargetFile
	//将没有匹配文件的 与 MD5码与文件不同的目录找出
	Dir := CsDir.ContrastDirMD5(LocalNow.FileMD5, Backup.FileMD5, LocalNow.DirHead)
	printClientLog(id, "正在将需要新建的文件发给远端")
	WriteAgreement(conn, CsDir.PackSliceString(Dir)) //将没有匹配文件的 与 MD5码与文件不同的目录找出发给远端

	var dirName string
	for {
		dast, _ := ReadAgreement(conn)                        //接收数据
		if string(dast) == "The transfer file is finished!" { //是否已经发送完毕
			printClientLog(id, "接收完毕，关闭连接")
			break
		}

		dirName = CsDir.JointDir(SrcDir, string(dast))
		printClientLog(id, "正在接收需要创建的文件数据,文件：%s", dirName)
		buff, _ := ReadAgreement(conn) //接收数据
		printClientLog(id, "正在创建文件,文件：%s", dirName)
		CsDir.WriteFileAll(dirName, buff)
		printClientLog(id, "创建成功，文件：%s", dirName)
	}
}
