package logging

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const dateString = "06-01-02::03:04:05"

// is pointing to to file and carries all LogFile methods
type LogFile struct {
	name        string
	path        string
	maxLines    uint16 // max 65535 // with 100 chars per line 10 000 lines ~= 1mb // 0 ==> max
	lines       uint16
	flag        int
	logger      *log.Logger
	currentFile string
}

// creates new file if necessary and returns new LogFile, if maxLines is 0, defaults to 65535
func New(name, path string, maxLines, flag int) (LogFile, error) {
	if maxLines > math.MaxUint16 || maxLines == 0 {
		maxLines = math.MaxUint16
	}
	l := LogFile{name: name, path: path, lines: 0, maxLines: uint16(maxLines), flag: flag}
	err := l.new()
	return l, err
}
func (l *LogFile) Log(messages ...interface{}) {
	l.update()
	s := fmt.Sprintln(messages...)
	l.logger.Print(s)
}
func (l *LogFile) Error(messages ...interface{}) {
	l.update()
	s := "error:"
	s += fmt.Sprintln(messages...)
	l.logger.Print(debugInfo() + s)
}
func (l *LogFile) Panic(messages ...interface{}) {
	l.update()
	s := "PANIC:\n         "
	s += fmt.Sprintln(messages...)
	l.logger.Panic(debugInfo() + s)
}
func (l *LogFile) Fatal(messages ...interface{}) {
	l.update()
	s := "FATAL:\n         "
	s += fmt.Sprintln(messages...)
	l.logger.Fatal(debugInfo() + s)
}
func (l *LogFile) update() {
	l.lines++
	if fileExists(l.currentFile) && l.lines < l.maxLines-1 {
		return
	}
	err := l.new()
	if err != nil {
		panic(err)
	}
}

func (l *LogFile) new() error {
	fis, err := ioutil.ReadDir(l.path)
	if err != nil {
		return err
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
			if lines < int(l.maxLines) {
				l.lines = uint16(lines)
				l.currentFile = filepath.Join(l.path, fi.Name())
				f, err := os.OpenFile(l.currentFile, os.O_APPEND|os.O_WRONLY, 0644)
				l.logger = log.New(f, "", l.flag)
				return err
			}
			err = f.Close()
			if err != nil {
				return err
			}
		}
	}
	l.currentFile = fmt.Sprint(l.name, time.Now().Format(dateString), ".log")
	f, err := os.Create(filepath.Join(l.path, l.currentFile))
	if err != nil {
		return err
	}
	l.logger = log.New(f, "", l.flag)
	l.Log("file started", time.Now().String())
	return err
}

// https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func debugInfo() string {
	var info string
	caller := getFrame(2)
	callersCaller := getFrame(3)
	info += fmt.Sprintln("")
	info += fmt.Sprintln("current:", cutSrcPath(caller.Function), cutSrcPath(caller.File), caller.Line)
	info += fmt.Sprintln("caller:", cutSrcPath(callersCaller.Function), cutSrcPath(callersCaller.File), callersCaller.Line)
	return info
}
func cutSrcPath(s string) string {
	cutset := string(filepath.Separator) + "src" + string(filepath.Separator) //src as is common in go
	if strings.Contains(s, cutset) {
		i := strings.Index(s, cutset)
		return s[i+len(cutset):]
	}
	return s
}

func getFrame(skipFrames int) runtime.Frame {
	// We need the frame at index skipFrames+2, since we never want runtime.Callers and getFrame
	targetFrameIndex := skipFrames + 2
	// Set size to targetFrameIndex+2 to ensure we have room for one more caller than we need
	programCounters := make([]uintptr, targetFrameIndex+2)
	n := runtime.Callers(0, programCounters)
	frame := runtime.Frame{Function: "unknown"}
	if n > 0 {
		frames := runtime.CallersFrames(programCounters[:n])
		for more, frameIndex := true, 0; more && frameIndex <= targetFrameIndex; frameIndex++ {
			var frameCandidate runtime.Frame
			frameCandidate, more = frames.Next()
			if frameIndex == targetFrameIndex {
				frame = frameCandidate
			}
		}
	}
	return frame
}
