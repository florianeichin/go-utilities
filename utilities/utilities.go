package utilities

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/logrusorgru/aurora"
)

/*
Printer
Logger for Writers/Readers
*/

// NewPrinter finds out who is calling, which function and which file. acepteed values: information to print, type (FUNC, DEBU, ERR)
func NewPrinter(dataType, comment string, content ...interface{}) {
	var formattedDataType aurora.Value
	caller := GetCaller(3)
	switch dataType {
	case "DEBUG":
		if strings.Contains(caller, "vendor") {
			formattedDataType = aurora.Magenta(fmt.Sprintf("[%s] %s() %s: %v", dataType, caller, comment, content))
		} else {
			formattedDataType = aurora.Green(fmt.Sprintf("[%s] %s() %s: %v", dataType, caller, comment, content))
		}

	case "ERROR":
		formattedDataType = aurora.Red(fmt.Sprintf("[%s] %s() %s:  %v", dataType, caller, comment, content))
	case "FUNCT":
		if strings.Contains(caller, "vendor") {
			formattedDataType = aurora.Magenta(fmt.Sprintf("[%s] %s()", dataType, caller))
		} else {
			formattedDataType = aurora.Blue(fmt.Sprintf("[%s] %s()", dataType, caller))
		}
	case "CALL":
		formattedDataType = aurora.Gray(fmt.Sprintf("[%s] %s() %s:  %v", dataType, caller, comment, content))
	}

	fmt.Printf("%s\n", formattedDataType)

}
func PrintCallers(steps int) {
	for i := 3; i < steps+2; i++ {
		NewPrinter("DEBUG", "CALL", GetCaller(i))
	}
}

func GetCaller(steps int) string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(steps, fpcs)
	if n == 0 {
		return "n/a"
	}
	fun := runtime.FuncForPC(fpcs[0] - 1)
	if fun == nil {
		return "n/a"
	}
	return fun.Name()
}

func NewWriterLogger(w io.Writer, path string) io.Writer {
	file, err := os.Create(path)
	if err != nil {
		NewPrinter("ERROR", "", err)
	}
	mw := io.MultiWriter(file, w)
	return mw
}

func NewWriteCloserLogger(w io.Writer, path string) io.WriteCloser {
	return &writeCloserLogger{NewWriterLogger(w, path)}
}

type writeCloserLogger struct {
	io.Writer
}

func (wcl *writeCloserLogger) Close() error {
	return nil
}

func NewReaderLogger(r io.Reader, path string) io.Reader {
	file, err := os.Create(path)
	if err != nil {
		NewPrinter("ERROR", "", err)
	}
	tee := io.TeeReader(r, file)
	return tee
}
