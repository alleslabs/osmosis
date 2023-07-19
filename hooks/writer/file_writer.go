package writer

import (
	"context"
	"fmt"
	"os"

	"github.com/osmosis-labs/osmosis/v16/hooks/common"
)

type FileWriter struct {
	file *os.File
}

func NewFileWrite(dir string) *FileWriter {
	f, err := os.OpenFile("messages.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	return &FileWriter{file: f}
}

func (f FileWriter) WriteMessages(ctx context.Context, msgs []common.Message) {
	for _, msg := range msgs {
		_, err := f.file.Write([]byte(fmt.Sprintf("key:%s, value:%s\n", msg.Key, msg.Value)))
		if err != nil {
			panic(err)
		}
	}
}
