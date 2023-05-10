package comfy_log

import "os"

type logFileWriter struct {
	// writeToStdoutFile 为 true 时, 该 logFileWriter 将会向 stdoutFile 写入日志. 否则, 将会向 stderrFile 写入日志.
	writeToStdoutFile bool
}

func (w logFileWriter) Write(p []byte) (n int, err error) {
	var file *os.File
	if w.writeToStdoutFile {
		file = stdoutFile
	} else {
		file = stderrFile
	}

	return file.Write(p)
}
