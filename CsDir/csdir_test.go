package CsDir

import (
	"fmt"
	"testing"
)

func TestCsDir(t *testing.T) {
	SrcDir := "E:/golang/csfioletext/go"
	//ShareDir := "E:/golang/csfioletext/Share"
	ShareDir := "./"
	Suffix := ""

	var i_Walkdir Walkdir_i
	var s_walkdir Walkdir_s
	i_Walkdir = &s_walkdir
	s_walkdir.WalkDirInit(SrcDir, ShareDir, Suffix)
	i_Walkdir.WalkDirFile()
	err := i_Walkdir.MakeDirs()
	if err != nil {
		t.Error("error!\n")
	} else {
		fmt.Println("nice day!")
	}

}
