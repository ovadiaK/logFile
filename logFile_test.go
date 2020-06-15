package logging

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testFolder = "testFolder"
	testLog    = "testLog"
	numOfLogs  = 1
)

// check if new file is created when old is full
func TestNew(t *testing.T) {
	numLinesArray := []int{10, 2000, 20000, 65000, 0}
	for _, numLines := range numLinesArray {
		l := buildUp(numLines)
		for i := 0; i < numLines*numOfLogs; i++ {
			l.Log("message ", i, "lines:", l.lines)
			if i%numLines == int(l.maxLines)-(numLines/10) {
				time.Sleep(time.Second)
			}
		}
		fis, err := ioutil.ReadDir(l.path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fi := range fis {
			if strings.Contains(fi.Name(), l.name) {
				if fi.IsDir() {
					continue
				}
				f, err := os.Open(filepath.Join(l.path, fi.Name()))
				if err != nil {
					continue
				}
				lines, err := lineCounter(f)
				if lines > int(l.maxLines) {
					t.Fatal("expected:", l.maxLines, "got:", lines)
				} else {
					fmt.Printf("expected: %v got %v\n", l.maxLines, lines)
				}
			}
		}
		tearDown()
	}
}

// check debug info
func TestLogFile_Error(t *testing.T) {
	l := buildUp(2000)
	func1 := func() {
		l.Error("func1")
		return
	}
	func2 := func() {
		l.Error("func2")
		func1()
		return
	}
	func2()
	fis, err := ioutil.ReadDir(l.path)
	if err != nil {
		t.Fatal(err)
	}
	for _, fi := range fis {
		if strings.Contains(fi.Name(), l.name) {
			if fi.IsDir() {
				continue
			}
			f, err := os.Open(filepath.Join(l.path, fi.Name()))
			if err != nil {
				continue
			}
			var s string
			if _, err := f.Read([]byte(s)); err != nil {
				t.Fatal(err)
			}

		}
	}
	//t.Fatal("end")
	//tearDown()
}

func buildUp(numLines int) LogFile {
	if err := os.Mkdir(testFolder, os.ModePerm); err != nil {
		//t.Fatal(err)
		fmt.Println(err)
	}
	l, err := New(testLog, testFolder, numLines, 3)
	if err != nil {
		//t.Fatal(err)
		fmt.Println(err)
	}
	return l
}
func tearDown() {
	if err := os.RemoveAll(testFolder); err != nil {
		fmt.Println(err)
	}
}
