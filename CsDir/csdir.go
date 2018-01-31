package CsDir

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Walkdir_i interface {
	//WalkDirInit(SrcDir string, BuildDir string, Suffix string)
	WalkDirFile() (err error)
	MakeDirs() (err error)
}

type Walkdir_s struct {
	srcDir     string   //源目录名
	buildDir   string   //需要新建备份文件的目录名
	suffix     string   //按文件后缀查找文件
	Files      []string //包含文件的目录名
	Dirs       []string //所有的目录名
	TargetDir  []string //不包含文件的目标文件夹
	TargetFile []string //不包含文件的目标文件夹
	FileMD5    []string //包含MD5码的文件目录，格式为MD5+TargetFile
	DirHead    string   //除了要共享的文件外的目录头
}

//获取指定目录及所有子目录下的所有文件与所有目录，可以匹配后缀过滤。
func (walkdir *Walkdir_s) WalkDirFile(SrcDir string, BuildDir string, Suffix string) (err error) {
	walkdir.srcDir = SrcDir
	walkdir.buildDir = BuildDir
	walkdir.suffix = Suffix
	walkdir.DirHead = getDirHead(walkdir.srcDir)

	ok, err := PathExists(walkdir.srcDir) //判断需要遍历的目录是否存在
	// if err != nil { //忽略错误
	// 	fmt.Println(err)
	// }
	if ok { //目录存在
		fmt.Println("Path Exists!")
	} else { //目录不存在
		fmt.Println("Path not exist!")
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

	for index, value := range walkdir.Dirs {
		if index != 0 {
			walkdir.TargetDir = append(walkdir.TargetDir, GetTargetDir(value, walkdir.DirHead))
		}
	}

	for _, value := range walkdir.Files {
		walkdir.TargetFile = append(walkdir.TargetFile, GetTargetDir(value, walkdir.DirHead))
	}

	for _, value := range walkdir.Files {
		walkdir.FileMD5 = append(walkdir.FileMD5, PackFileMD5(value))
	}
	return
}

func PackFileMD5(name string) string {
	var Md5 [16]byte
	var buffer bytes.Buffer

	Md5 = GetMD5(name)
	md5 := Md5[:]
	buffer.Write(md5)
	buffer.WriteString(name)
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

//将[]string里的目录，去掉DirHead相对路径后在，共享文件夹ShareDir创建出来
func (walkdir *Walkdir_s) MakeDirs() (err error) {
	for index, value := range walkdir.Dirs {
		//fmt.Println("Index = ", index, "Value = ", value)
		if index != 0 {
			dir := GetShareDir(value, walkdir.buildDir, walkdir.DirHead)
			err = MakeDir(dir)
			if err != nil {
				return
			}
		}
	}
	return
}

//创建目录
func MakeDir(name string) (err error) {
	err = os.MkdirAll(name, 0777)
	if err != nil {
		fmt.Printf("%s\n", err)
	} else {
		// fmt.Print("Create Directory OK!\n")
	}
	return
}

//将共享路径头提取出来
func getDirHead(name string) (DirHead string) {
	catstring := filepath.Dir(name) //filepath.Dir可以将最后一个文件夹去掉
	// fmt.Println("bi", name)
	// fmt.Println("bo", catstring)
	srcString := []byte(name)
	catByte := []byte(catstring)
	DirLen := len(catByte) + 1 //加1的目的是去掉'/'
	DirHead = string(srcString[:DirLen])
	return
}

/*将需要共享的文件名转换成特定格式
SrcDir 共享的文件名,绝对路径
BuildDir 共享目录
DirHead 除了要共享的文件夹外，前面的性对路径头

Dir 返回值 是要共享的目录的纯目录
*/
func GetShareDir(SrcDir string, BuildDir string, DirHead string) (Dir string) {
	var ShareDirNew string

	SrcDirByte := []byte(SrcDir)
	ShareDirByte := []byte(BuildDir)
	DirHeadLen := len(DirHead)

	DirTail := SrcDirByte[DirHeadLen:]

	if len(BuildDir) > 0 {
		if ShareDirByte[len(ShareDirByte)-1] != byte('/') { //如果共享文件夹最后一个字符不是'/'
			ShareDirNew = BuildDir + "/"
		} else {
			ShareDirNew = BuildDir
		}
	}
	Dir = ShareDirNew + string(DirTail)
	return
}

//获取目标目录
func GetTargetDir(SrcDir string, DirHead string) (Dir string) {
	SrcDirByte := []byte(SrcDir)
	DirHeadLen := len(DirHead)
	DirTail := SrcDirByte[DirHeadLen:]
	Dir = string(DirTail)
	return
}

//拼接文件夹
func JointDir(jointDir string, TargetDir string) (Dir string) {
	jointDirByte := []byte(jointDir)
	if jointDirByte[len(jointDirByte)-1] != byte('/') || jointDirByte[len(jointDirByte)-1] != byte('\\') { //如果共享文件夹最后一个字符不是'/'或'\'
		Dir = jointDir + "/" + TargetDir
	} else {
		Dir = jointDir + TargetDir
	}
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

//判断文件或目录是否存在
//如果路径存在返回false，不存在返回nil
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

//本地的文件夹初始化
func DirInitLocal() (SrcDir string, BuildDir string, Suffix string) {
	SrcDir = "E:/golang/gopath/src/github.com/CsBoBoNice/Local"
	BuildDir = ""
	Suffix = ""
	return
}

//远端的文件夹初始化
func DirInitRemote() (SrcDir string, BuildDir string, Suffix string) {
	SrcDir = "E:/golang/gopath/src/github.com/CsBoBoNice/Local"
	BuildDir = ""
	Suffix = ""
	return
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
