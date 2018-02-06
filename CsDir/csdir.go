package CsDir

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Walkdir_i interface {
	WalkDirFile() (err error)
	MakeDirs() (err error)
}

type Walkdir_s struct {
	srcDir string //源目录名
	// buildDir   string   //需要新建备份文件的目录名
	suffix     string   //按文件后缀查找文件
	Files      []string //包含文件的目录名
	Dirs       []string //所有的目录名
	TargetDir  []string //不包含文件的目标文件夹
	TargetFile []string //不包含文件的目标文件夹
	FileMD5    []string //包含MD5码的文件目录，格式为MD5+TargetFile
	DirHead    string   //除了要共享的文件外的目录头
}

//获取指定目录及所有子目录下的所有文件与所有目录，可以匹配后缀过滤。
func (walkdir *Walkdir_s) WalkDirFile(SrcDir string, Suffix string) (err error) {
	walkdir.srcDir = SrcDir
	// walkdir.buildDir = BuildDir
	walkdir.suffix = Suffix
	walkdir.DirHead = GetDirHead(walkdir.srcDir)

	ok, err := PathExists(walkdir.srcDir) //判断需要遍历的目录是否存在
	// if err != nil { //忽略错误
	// 	fmt.Println(err)
	// }
	if ok { //目录存在
		// fmt.Printf("%s目录存在!\n", walkdir.srcDir)
	} else { //目录不存在
		fmt.Printf("\t%s目录不存在!\n", walkdir.srcDir)
		MakeDir(walkdir.srcDir) //目录不存在则创建目录
		return                  //目录不存在，遍历目录就没有必要了，直接返回
	}

	//遍历目录
	walkdir.Files = make([]string, 0, 30)
	walkdir.Dirs = make([]string, 0, 30)
	walkdir.suffix = strings.ToUpper(walkdir.suffix)                                             //忽略后缀匹配的大小写
	err = filepath.Walk(walkdir.srcDir, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		//if err != nil { //忽略错误
		// return err
		//}
		if fi.IsDir() {
			walkdir.Dirs = append(walkdir.Dirs, filename)
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), walkdir.suffix) {
			walkdir.Files = append(walkdir.Files, filename)
			return nil
		}
		return nil
	})

	for _, value := range walkdir.Dirs {
		// if index != 0 {
		walkdir.TargetDir = append(walkdir.TargetDir, GetTargetDir(value, walkdir.DirHead))
		// }
	}

	for _, value := range walkdir.Files {
		walkdir.TargetFile = append(walkdir.TargetFile, GetTargetDir(value, walkdir.DirHead))
	}

	for _, value := range walkdir.Files {
		walkdir.FileMD5 = append(walkdir.FileMD5, PackFileMD5(value, walkdir.DirHead))
	}
	return
}

func ByteToByte(add []byte, too []byte) []byte {
	for _, v := range too {
		add = append(add, v)
	}
	return add
}

func PackSliceString(buff []string) []byte {
	var buffer bytes.Buffer
	var TotalNum uint64
	var lenNum uint64
	TotalNum = uint64(len(buff))

	buffer.Write(Uint64ToByte(TotalNum))

	for _, v := range buff {
		lenNum = uint64(len(v))
		buffer.Write(Uint64ToByte(lenNum))
		// fmt.Printf("每条数据长度：%d\n", lenNum)
	}
	for _, v := range buff {
		buffer.Write([]byte(v))
	}
	// Date = buffer.Bytes()
	return buffer.Bytes()
}

func UnpackSliceString(buff []byte) (SliceString []string) {
	var TotalNum uint64  //string切片总数
	var lenNum uint64    //每条string字节数
	var scaler uint64    //循环计数器
	var LenBuff []uint64 //每条string字节数数组
	var bufflen uint64   //总数据长度
	var headLen uint64   //非真正数据片段的数据头
	scaler = 0
	bufflen = uint64(len(buff))
	TotalNum = ByteToUint64(buff)
	headLen = (TotalNum + 1) * 8 //因为数据头每个数据都是uint64,64位为8个字节
	if bufflen < headLen {       //如果总数据比 每条string字节数数组 数据还要少return
		return
	}
	// fmt.Printf("数组总数：%d\n", TotalNum)
	LenbuffByte := buff[8 : 8*(TotalNum+1)]  //将只含有每条string字节数的数据片段切割出来
	SliceBuff := buff[(8 * (TotalNum + 1)):] //将含有真正的数据片段切割出来
	for i := 0; i < int(TotalNum); i++ {
		lenNum = ByteToUint64(LenbuffByte[i*8:])
		LenBuff = append(LenBuff, lenNum)
		scaler = scaler + lenNum
		// fmt.Printf("每条数据长度：%d\n", lenNum)
	}
	if bufflen != scaler+headLen { //如果总数据字节数和要解析的字节数不符，return
		return
	}
	scaler = 0
	for _, v := range LenBuff {
		SliceString = append(SliceString, string(SliceBuff[scaler:scaler+v]))
		scaler = scaler + v
	}

	return
}

func ByteToUint64(date []byte) (i uint64) {
	i = binary.BigEndian.Uint64(date[0:8])

	// fmt.Println(i)
	return
}

func Uint64ToByte(i uint64) (date []byte) {
	date = make([]byte, 8)
	binary.BigEndian.PutUint64(date, uint64(i))
	return
}

func PackFileMD5(SrcDir string, DirHead string) string {
	var Md5 [16]byte
	var buffer bytes.Buffer

	Md5 = GetMD5(SrcDir)
	md5 := Md5[:]
	buffer.Write(md5)
	buffer.WriteString(GetTargetDir(SrcDir, DirHead))
	return buffer.String()
}

func UnpackFileMD5(data string) ([16]byte, string) {
	var buffer bytes.Buffer
	buffer.WriteString(data)
	var Md5 [16]byte
	for i := 0; i < 16; i++ {
		md5, err := buffer.ReadByte()
		if err != nil {
			return Md5, ""
		}
		Md5[i] = md5
	}
	return Md5, buffer.String()
}

/*
	将共享路径头提取出来
例：
	name = /abc/123/def/321
	DirHead = /abc/123/def
*/
func GetDirHead(name string) (DirHead string) {
	catstring := filepath.Dir(name) //filepath.Dir可以将最后一个文件夹去掉
	// fmt.Println("bi", name)
	// fmt.Println("bo", catstring)
	srcString := []byte(name)
	catByte := []byte(catstring)
	DirLen := len(catByte) + 1 //加1的目的是去掉'/'
	DirHead = string(srcString[:DirLen])
	return
}

/*
	获取目标目录
例：
	jointDir = /abc/123
	TargetDir = /def/321
	Dir = /abc/123/def/321
*/
func GetTargetDir(SrcDir string, DirHead string) (Dir string) {
	SrcDirByte := []byte(SrcDir)
	DirHeadLen := len(DirHead)
	DirTail := SrcDirByte[DirHeadLen:]
	Dir = string(DirTail)
	return
}

/*
	拼接文件夹
例：
	jointDir = /abc/123
	TargetDir = /def/321
	Dir = /abc/123/def/321
*/
func JointDir(jointDir string, TargetDir string) (Dir string) {
	jointDirByte := []byte(jointDir)
	var va byte = '\\'
	if jointDirByte[len(jointDirByte)-1] != byte('/') || jointDirByte[len(jointDirByte)-1] != va { //如果共享文件夹最后一个字符不是'/'或'\'
		Dir = jointDir + "/" + TargetDir
	} else {
		Dir = jointDir + TargetDir
	}
	return
}

//将前一个目录与后一个的最后一个目录拼接
//拼接文件夹
/*
例：
	jointDir = /abc/123
	TargetDir = /def/321/aaa
	Dir = /abc/123/aaa
*/
func JointDir2(jointDir string, TargetDir string) (Dir string) {
	targetDir := GetTargetDir(TargetDir, GetDirHead(TargetDir))
	Dir = JointDir(jointDir, targetDir)
	return
}

//将整个文件读取出来得到文件数据的MD5码
func GetMD5(name string) (MD5Byte [16]byte) {
	f, err := os.Open(name)
	if err != nil {
		fmt.Println("Open", err)
		return
	}
	defer f.Close()
	body, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("ReadAll", err)
		return
	}
	MD5Byte = md5.Sum(body)
	return
}

//得到文件的全部字节
func ReadFileAll(name string) (date []byte) {
	f, err := os.Open(name)
	if err != nil {
		fmt.Println("Open", err)
		return
	}
	defer f.Close()
	date, err = ioutil.ReadAll(f)
	if err != nil {
		fmt.Println("ReadAll", err)
		return
	}
	return
}

//函数功能：写入文件
//参数：1，写入的文件名，2，写入的数据 3，写入的字节数
//返回值：1，是否出错
func WriteFileAll(name string, buff []byte) (err error) {

	var num int
	fo, err := os.Create(name) //创建输出*File 写文件
	if err != nil {
		panic(err)
	}
	defer fo.Close() //退出后关闭文件

	fmt.Printf("\t写入到%s\t", name)
	num, err = fo.Write(buff)
	if err != nil { //写入output.txt,直到错误 写文件
		panic(err)
	}
	fmt.Printf("写入大小=%fMB\n", float64(num)/1024/1024)

	return
}

//判断文件或目录是否存在
//如果路径存在返回true，不存在返回false
//不知道路径是否存在err!=nil
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

//判断是否是文件夹
//如果是文件夹返回true，不存在返回false
//不知道路径是否存在err!=nil
func IsDir(path string) (bool, error) {
	ok, err := PathExists(path) //先判断文件或目录是否存在
	if ok != true {
		return false, err
	}
	f, _ := os.Stat(path) //再判断是否是文件夹
	isDir := f.IsDir()
	err = nil
	return isDir, err
}

//创建目录
func MakeDir(name string) (err error) {
	err = os.MkdirAll(name, 0777)
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("创建目录%s\n", name)
	}
	return
}

//删除文件或目录
func DeleteDir(name string) (err error) {
	os.RemoveAll(name)
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		fmt.Printf("删除文件%s\n", name)
	}
	return
}

//对比本地目录与远端目录，以发送过来的远端目录为基准，将多余的，目录删除，不足的目录新建
func ContrastDir(Local []string, Backup []string, DirHead string) {
	var identical []string
	for _, v := range Backup {
		for j, va := range Local {
			if va == v {
				Local = append(Local[:j], Local[j+1:]...) //将相同的目录删除，这样本地目录就只剩下需要删除的目录了
				identical = append(identical, v)          //相同目录的保存到identical中
				break
			}
		}
	}
	for _, v := range identical {
		for j, va := range Backup {
			if va == v {
				Backup = append(Backup[:j], Backup[j+1:]...) //将本地目录与相同目录切片里的数据去掉，这样远程目录就只剩下需要创建的目录了
				break
			}
		}
	}
	for _, v := range Local {
		DeleteDir(JointDir(DirHead, v))
	}
	for _, v := range Backup {
		MakeDir(JointDir(DirHead, v))
	}
}

func ContrastDirMD5(Local []string, Backup []string, DirHead string) (Dir []string) {
	var identical []string
	// var md5 [16]byte
	var dir string
	for _, v := range Backup {
		for j, va := range Local {
			if va == v {
				Local = append(Local[:j], Local[j+1:]...) //将相同的目录去掉，这样本地目录就只剩下需要删除的目录了
				identical = append(identical, v)          //相同目录的保存到identical中
				break
			}
		}
	}
	for _, v := range identical {
		for j, va := range Backup {
			if va == v {
				Backup = append(Backup[:j], Backup[j+1:]...) //将本地目录与相同目录切片里的数据去掉，这样远程目录就只剩下需要创建的目录了
				break
			}
		}
	}
	for _, v := range Local {
		_, dir = UnpackFileMD5(v)
		DeleteDir(JointDir(DirHead, dir))
	}

	for _, v := range Backup {
		_, dir = UnpackFileMD5(v)
		Dir = append(Dir, dir) //将需要远端发送过来的目录保存到Dir里
	}
	return
}

//本地的文件夹初始化
func DirInitLocal() (Local string, Suffix string) {
	Local = ""
	Suffix = ""
	return
}

//远端的文件夹初始化
func DirInitRemote() (Local string, Backup string, Suffix string) {
	Local = "F:/Test/Backup"
	// Backup = "F:/Test/1234567"
	Backup = "F:/Test/Local"
	Suffix = ""
	return
}

func ListMD5File(p []string) {
	for index, value := range p {
		// fmt.Println("Index = ", index, "Value = ", value)
		fmt.Printf("I=%d\t%v\t%s\n", index, []byte(value)[:16], string([]byte(value[16:])))
	}
}

func ListFileFunc(p []string) {
	for index, value := range p {
		fmt.Println("Index = ", index, "Value = ", value)
	}
}

/*************下面的是网上粘下来的代码*******************************/
/*
func convert() {
	stringSlice := []string{"E:/golang/csfioletext"}

	stringByte := strings.Join(stringSlice, "\x20\x00") // x20 = space and x00 = null

	buff := []byte(stringByte)

	fmt.Println([]byte(stringByte))
	fmt.Println(buff[3:])
	bufftring := "./" + string(buff[3:])
	fmt.Println(string([]byte(stringByte)))
	fmt.Println(bufftring)
}

//获取指定目录下的所有文件，不进入下一级目录搜索，可以匹配后缀过滤。
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

//获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix)                                                     //忽略后缀匹配的大小写
	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		//if err != nil { //忽略错误
		// return err
		//}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}
*/
