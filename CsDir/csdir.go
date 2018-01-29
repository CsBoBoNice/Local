package CsDir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Walkdir_i interface {
	//WalkDirInit(SrcDir string, ShareDir string, Suffix string)
	WalkDirFile() (err error)
	MakeDirs() (err error)
}

type walkdir_s struct {
	srcDir   string   //源目录名
	shareDir string   //共享目录名（将要拷贝到的目录）
	suffix   string   //按文件后缀查找文件
	Files    []string //包含文件的目录名
	Dirs     []string //所有的目录名
	DirHead  string   //除了要共享的文件外的目录头
}

func (walkdir *walkdir_s) WalkDirInit(SrcDir string, ShareDir string, Suffix string) {
	walkdir.srcDir = SrcDir
	walkdir.shareDir = ShareDir
	walkdir.suffix = Suffix
	walkdir.DirHead = getDirHead(walkdir.srcDir)

}

//获取指定目录及所有子目录下的所有文件与所有目录，可以匹配后缀过滤。
func (walkdir *walkdir_s) WalkDirFile() (err error) {
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
	return
}

//将[]string里的目录，去掉DirHead相对路径后在，共享文件夹ShareDir创建出来
func (walkdir *walkdir_s) MakeDirs() (err error) {
	for index, value := range walkdir.Dirs {
		//fmt.Println("Index = ", index, "Value = ", value)
		if index != 0 {
			dir := GetShareDir(value, walkdir.shareDir, walkdir.DirHead)
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
ShareDir 共享目录
DirHead 除了要共享的文件夹外，前面的性对路径头

Dir 返回值 是要共享的目录的纯目录
*/
func GetShareDir(SrcDir string, ShareDir string, DirHead string) (Dir string) {
	var ShareDirNew string

	SrcDirByte := []byte(SrcDir)
	ShareDirByte := []byte(ShareDir)
	DirHeadLen := len(DirHead)

	DirTail := SrcDirByte[DirHeadLen:]

	if len(ShareDir) > 0 {
		if ShareDirByte[len(ShareDirByte)-1] != byte('/') { //如果共享文件夹最后一个字符不是'/'
			ShareDirNew = ShareDir + "/"
		} else {
			ShareDirNew = ShareDir
		}
	}
	Dir = ShareDirNew + string(DirTail)
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
