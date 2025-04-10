// package dbline implements a very basic database based on file lines append
// types with two methods. String to create the line and Parse to create a type
// from a line
package dbline

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

type dbStructElemPointer[T any] interface {
	dbStructElem
	*T
}

type dbStructElem interface {
	String() string
	Parse(s string) error
}

func DeleteEntry(path string, match string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	buf := new(bytes.Buffer)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.Contains(text, match) {
			_, err := buf.Write(scanner.Bytes())
			if err != nil {
				return err
			}
			_, err = buf.WriteString("\n")
			if err != nil {
				return err
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	return os.WriteFile(f.Name(), buf.Bytes(), 0600)
}

func AppendEntry[T dbStructElem](path string, item T) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(item.String() + "\n"); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}

func Save[T dbStructElem](path string, data []T) error {
	buf := new(bytes.Buffer)
	for _, v := range data {
		buf.WriteString(v.String() + "\n")
	}
	os.WriteFile(path, buf.Bytes(), 0600)
	return nil
}

func Open[P dbStructElemPointer[T], T any](path string) ([]T, error) {
	var ret []T
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var e P = new(T)
		err := e.Parse(scanner.Text())
		if err != nil {
			return nil, err
		}
		ret = append(ret, *e)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}
