package main

import (
	"io"
)

type walDB struct {
	file io.ReadWriteSeeker

	Term  byte
	Bsize int
}

func (fl *walDB) SetWal(key, value []byte) error {

	entry := make([]byte, fl.Bsize)

	keyLen, valueLen := len(key), len(value)

	copy(entry[:4], []byte("set "))
	copy(entry[4:keyLen+4], key)
	copy(entry[keyLen+4:keyLen+5], []byte(" "))
	copy(entry[keyLen+5:keyLen+valueLen+5], value)

	// Writing padding
	for i := keyLen + valueLen + 5; i < fl.Bsize-1; i++ {
		entry[i] = '#'
	}
	entry[fl.Bsize-1] = '\n'

	if _, err := fl.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	if _, err := fl.file.Write(entry); err != nil {
		return err
	}
	return nil
}

func (fl *walDB) DelWal(key []byte) error {

	entry := make([]byte, fl.Bsize)

	keyLen := len(key)

	copy(entry[:4], []byte("del "))
	copy(entry[4:keyLen+4], key)

	// Writing padding
	for i := keyLen + 4; i < fl.Bsize-1; i++ {
		entry[i] = '#'
	}
	entry[fl.Bsize-1] = '\n'

	if _, err := fl.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	if _, err := fl.file.Write(entry); err != nil {
		return err
	}
	return nil

}

func (fl *fileDB) WriteOnEnd(valueToWrite []byte) error {
	if _, err := fl.file.Seek(0, io.SeekEnd); err != nil {
		return err
	}
	if _, err := fl.file.Write(valueToWrite); err != nil {
		return err
	}
	return nil
}

func NewwalDB(f io.ReadWriteSeeker) *walDB {
	return &walDB{
		file:  f,
		Term:  byte('#'),
		Bsize: 100,
	}
}
