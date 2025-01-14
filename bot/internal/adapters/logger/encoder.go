package logger

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

type prefixEncoder struct {
	// embed a zapcore encoder
	// this makes prependEncoder implement the interface without extra work
	zapcore.Encoder

	// zap buffer pool
	pool buffer.Pool

	prefix string
}

func (e *prefixEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	// new log buffer
	buf := e.pool.Get()

	// prepend the JournalD prefix based on the entry level
	buf.AppendString(e.prefix)
	buf.AppendString(" ")

	// calling the embedded encoder's EncodeEntry to keep the original encoding format
	consolebuf, err := e.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	// just write the output into your own buffer
	_, err = buf.Write(consolebuf.Bytes())
	if err != nil {
		return nil, err
	}
	return buf, nil
}
