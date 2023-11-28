package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type fileDB struct {
	file        io.ReadWriteSeeker
	noFiles     int
	maxFileSize int
}

func newFileDB(sz int) (*fileDB, error) {

	dir, _ := filepath.Split("sst_1.sst")
	files, err := filepath.Glob(filepath.Join(dir, "sst_*.sst"))
	if err != nil {
		return nil, err
	}

	noFiles := len(files)
	var f *os.File
	f, _ = os.OpenFile("testfile.txt", os.O_RDWR|os.O_CREATE, 0666)

	return &fileDB{
		file:        f,
		noFiles:     noFiles,
		maxFileSize: sz,
	}, nil
}

func (fl *fileDB) getFileVariable(s string) ([]byte, error) {
	offset := 0
	switch s {
	case "magicnumber":
		offset += 0
	case "entrycount":
		offset += 4
	case "smallestkey":
		offset += 8
	case "largestkey":
		offset += 12
	default:
		return nil, nil
	}
	if _, err := fl.file.Seek(int64(offset), io.SeekStart); err != nil {
		return nil, err

	}

	resBuffer := make([]byte, 4)
	_, err := fl.file.Read(resBuffer)
	if err != nil {
		return nil, err
	}

	return resBuffer, nil
}

func (fl *fileDB) SetVariable(s string, valueToSet []byte) error {
	offset := 0
	switch s {
	case "magicnumber":
		offset += 0
	case "entrycount":
		offset += 4
	case "smallestkey":
		offset += 8
	case "largestkey":
		offset += 12
	default:
		return nil
	}
	if _, err := fl.file.Seek(int64(offset), io.SeekStart); err != nil {
		return err
	}
	if _, err := fl.file.Write(valueToSet); err != nil {
		return err
	}

	return nil

}

func convertIntTo4ByteArray(n int) []byte {
	res := make([]byte, 4)
	binary.LittleEndian.PutUint32(res, uint32(n))
	return res
}

func convert4ByteArrayToInt(byteArray []byte) int {
	return int(binary.LittleEndian.Uint32(byteArray))
}
func (fl *fileDB) IncrementCount() error {
	currentCounter, err := fl.getFileVariable("entrycount")
	if err != nil {
		return err
	}
	num := binary.LittleEndian.Uint32(currentCounter)

	// Increment the integer
	num++

	// Convert the incremented integer back to a byte array
	binary.LittleEndian.PutUint32(currentCounter, num)

	if err := fl.SetVariable("entrycount", currentCounter); err != nil {
		return err
	}

	return nil

}
func (file *fileDB) createSST(entries map[string]string) error {
	// Seek to the end of the file to append
	_, err := file.file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	// Write SST file header
	header := make([]byte, magicNumberSize+entryCountSize+keyLengthSize+keyLengthSize)
	binary.BigEndian.PutUint32(header[:magicNumberSize], 12345) // Placeholder magic number
	binary.BigEndian.PutUint32(header[magicNumberSize:magicNumberSize+entryCountSize], uint32(len(entries)))
	binary.BigEndian.PutUint32(header[magicNumberSize+entryCountSize:magicNumberSize+entryCountSize+keyLengthSize], uint32(len(entries)))
	binary.BigEndian.PutUint32(header[magicNumberSize+entryCountSize+keyLengthSize:magicNumberSize+entryCountSize+keyLengthSize+keyLengthSize], uint32(len(entries)))
	_, err = file.file.Write(header)
	if err != nil {
		return err
	}

	// Write entries to SST file
	for key, value := range entries {
		// Determine operation type based on key prefix
		var opType byte
		if len(key) > 3 && key[:3] == "set" {
			opType = 1 // 1 for set

		} else {
			opType = 0 // 0 for del
		}
		key = key[4:]

		// Write value length and value
		valueb := []byte(value)
		keyb := []byte(key)
		valueLen := uint32(len(valueb))
		valueLenBuf := make([]byte, valueLengthSize)
		binary.BigEndian.PutUint32(valueLenBuf, valueLen)

		_, err = file.file.Write(valueb)
		if err != nil {
			return err
		}
		_, err = file.file.Write(valueLenBuf)
		if err != nil {
			return err
		}

		// Write key length and key
		keyLen := uint32(len(key))
		keyLenBuf := make([]byte, keyLengthSize)
		binary.BigEndian.PutUint32(keyLenBuf, keyLen)

		_, err = file.file.Write(keyb)
		if err != nil {
			return err
		}
		_, err = file.file.Write(keyLenBuf)
		if err != nil {
			return err
		}

		// Write operation type

		_, err = file.file.Write([]byte{opType})
		if err != nil {
			return err
		}
	}

	return nil
}

func (mem *memDB) createNewSSTFile() error {
	// Create a new SST file
	file, err := os.Create(fmt.Sprintf("sst_%d.sst", mem.file.noFiles+1))
	if err != nil {
		return err
	}
	mem.file.file = file
	mem.file.noFiles++
	return nil
}

func printSSTFileContents(fileDB *fileDB) {
	file, err := os.Open(fmt.Sprintf("sst_%d.sst", fileDB.noFiles))
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fmt.Println("SST File Contents:")
	io.Copy(os.Stdout, file)
}

func (mem *memDB) CreateNewFile() error {
	mem.file.noFiles += 1
	f, err := os.OpenFile("f"+fmt.Sprint(mem.file.noFiles)+".sst", os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return err
	}

	mem.file.file = f

	magicNumber := []byte{0, 0, 0, 0}
	entryCount := []byte{0, 0, 0, 0}
	smallestKey := []byte{0, 0, 0, 0}
	largestKey := []byte{0, 0, 0, 0}

	mem.file.WriteOnEnd(magicNumber)
	mem.file.WriteOnEnd(entryCount)
	mem.file.WriteOnEnd(smallestKey)
	mem.file.WriteOnEnd(largestKey)

	return nil

}