package main

import (
	"bytes"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
)

func randomBool() bool {
	if rand.Int()%2 == 0 {
		return false
	}
	return true
}

func RandInt64() int64 {
	return rand.Int63()
}

func RandStr(size int) string {
	letters := "abcdefghABCDEFGH0123456789"
	arr := strings.Split(letters, "")
	buf := bytes.Buffer{}
	for i := 0; i < size; i++ {
		id := rand.Intn(len(arr))
		buf.WriteString(arr[id])
	}
	return buf.String()
}

func JoinErrors(errs ...error) error {
	var sb strings.Builder
	for _, err := range errs {
		if sb.Len() > 0 {
			sb.WriteString(", and ")
		}
		sb.WriteString(err.Error())
	}
	return errors.New(sb.String())
}

func ReadFile(path string) (string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CreateCertFiles(fnames, texts []string) (fileKey, fileCrt string, err error) {
	for i, name := range fnames {
		SaveToFile(name, texts[i])
	}
	return fnames[0], fnames[1], nil
}

func SaveToFile(fname, text string) error {
	err := ioutil.WriteFile(fname, []byte(text), 0644)
	if err != nil {
		return err
	}
	return nil
}

func DeleteFile(paths ...string) error {
	for _, path := range paths {
		var err = os.Remove(path)
		if err != nil {
			return err
		}
	}
	return nil
}
