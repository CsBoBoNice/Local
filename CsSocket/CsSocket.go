package CsSocket

import (
	_ "bufio"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"net"
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
	fmt.Println(Datas.DataHeadbuff)
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
	fmt.Println(i)
	return
}

func Uint64ToByte(i uint64) (date []byte) {
	date = make([]byte, 8)
	binary.BigEndian.PutUint64(date, uint64(i))
	return
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
