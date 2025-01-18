package ch13

import (
	"io"

	"go.uber.org/multierr"
)

// struct sustainedMultiwriter
type sustainedMultiwriter struct {
	writers []io.Writer
}

// func Write
func (mw *sustainedMultiwriter) Write(p []byte) (n int, err error) {
	// for each writer
	for _, w := range mw.writers {
		// write and get error
		i, wErr := w.Write(p)
		// add up to the result
		n += i
		// append error to the result
		err = multierr.Append(err, wErr)
	}

	return n, err
}

// func SustainedMultiwriter
func SustainedMultiwriter(writers ...io.Writer) io.Writer {
	// create empty sustainedMultiwriter
	mw := sustainedMultiwriter{
		writers: make([]io.Writer, 0, len(writers)),
	}

	// for each writer
	for _, w := range writers {
		if m, ok := w.(*sustainedMultiwriter); ok {
			// append all aggregated writers if it's multiwriter
			mw.writers = append(mw.writers, m.writers...)
		} else {
			// or just writer itself if it's a simple writer
			mw.writers = append(mw.writers, w)
		}
	}

	return &mw
}
