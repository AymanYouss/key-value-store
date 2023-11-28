package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type Cmd int

const (
	Get Cmd = iota
	Set
	Del
	Ext
	Unk
	Flush
	Init
	Test
)
const (
	magicNumberSize = 4
	entryCountSize  = 4
	keyLengthSize   = 4
	valueLengthSize = 4
)

type Error int

func (e Error) Error() string {
	return "Empty command"
}

const (
	Empty Error = iota
)

type memDB struct {
	memValues []byte
	wal       *walDB
	file      *fileDB
}

func (mem *memDB) updateMemDisk() error {
	if len(mem.memValues) > 20 {
		err := mem.FlushMemToSSTFile()
		if err != nil {
			return err
		}
	}

	return nil
}

func (mem *memDB) Set(key, value []byte) error {

	mem.SetMem(key, value)
	mem.wal.SetWal(key, value)
	err := mem.updateMemDisk()
	if err != nil {
		return err
	}

	return nil
}

func (mem *memDB) Get(key []byte) ([]byte, error) {
	// First, try to get from memory
	value, err := mem.GetMem(key)
	if err != nil {
		return value, errors.New("Key not found")
	}

	if value != nil {
		return value, nil
	}

	// If not found in memory, try to get from SST files
	return mem.GetSST(key)
}

func (mem *memDB) Del(key string) (string, error) {
	// Check if the key is in the memTable
	if mem.keyExists(key) {
		// Add "del key" entry to memTable
		delEntry := fmt.Sprintf("del %s\n", key)
		mem.memValues = append(mem.memValues, []byte(delEntry)...)
		err := mem.updateMemDisk()
		if err != nil {
			return "", err
		}
		// Return the value associated with the key
		return mem.getValue(key), nil
	} else {
		val, err := mem.GetSST([]byte(key))
		if err != nil {
			return "", errors.New("key not found")
		}
		delEntry := fmt.Sprintf("del %s\n", key)
		mem.memValues = append(mem.memValues, []byte(delEntry)...)
		err = mem.updateMemDisk()
		if err != nil {
			return "", err
		}
		// Return the value associated with the key
		return string(val), nil

	}

}

func (mem *memDB) keyExists(key string) bool {
	// Check if the key exists in the memTable
	keySearch := fmt.Sprintf("set %s", key)
	return strings.Contains(string(mem.memValues), keySearch)
}

func (mem *memDB) getValue(key string) string {
	// Get the value associated with the key from the memValues
	lines := strings.Split(string(mem.memValues), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "set "+key+" ") {
			parts := strings.Fields(line)
			if len(parts) == 3 {
				return parts[2]
			}
		}
	}
	return ""
}

func (mem *memDB) SetMem(key, value []byte) error {

	entry := make([]byte, 0)

	entry = append(entry, []byte("set ")...)
	entry = append(entry, key...)
	entry = append(entry, []byte(" ")...)
	entry = append(entry, value...)
	entry = append(entry, []byte("\n")...)

	mem.memValues = append(mem.memValues, entry...)

	return nil
}

func (mem *memDB) GetMem(key []byte) ([]byte, error) {
	entries := strings.Split(string(mem.memValues), "\n")
	for i := len(entries) - 1; i >= 0; i-- {
		entry := entries[i]
		if strings.HasPrefix(entry, "del "+string(key)) {
			// Key is deleted, return an error
			return nil, errors.New("key not found")
		} else if strings.HasPrefix(entry, "set "+string(key)+" ") {
			// Extract the value from the entry and return it
			parts := strings.Fields(entry)
			if len(parts) == 3 {
				return []byte(parts[2]), nil
			}
		}
	}

	// Key not found
	return nil, nil
}

func (mem *memDB) DelMem(key []byte) ([]byte, error) {
	// Use the Get method to check if the key exists
	v, err := mem.GetMem(key)
	if err != nil {
		return nil, err // Key not found
	}

	// Key exists, create a "del" entry
	entry := []byte("del " + string(key) + "\n")
	mem.memValues = append(mem.memValues, entry...)

	return v, nil
}

func NewInMem() *memDB {

	f, err := os.OpenFile("wal.txt", os.O_CREATE, 0755)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	wal := NewwalDB(f)
	memValues := make([]byte, 0)

	maxFileSize := 100

	flDB, _ := newFileDB(maxFileSize)

	return &memDB{
		memValues,
		wal,
		flDB,
	}
}

func (mem *memDB) GetSST(key []byte) ([]byte, error) {
	// Iterate over SST files in reverse order

	for fileIndex := mem.file.noFiles; fileIndex > 0; fileIndex-- {

		// Open the SST file
		sstFileName := fmt.Sprintf("sst_%d.sst", fileIndex)

		file, err := os.Open(sstFileName)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		// Read the header of the SST file
		header := make([]byte, magicNumberSize+entryCountSize+keyLengthSize+keyLengthSize)
		_, err = file.Read(header)
		if err != nil {
			return nil, err
		}

		fileInfo, err := os.Stat(sstFileName)

		idx := fileInfo.Size()
		for idx > 16 {
			// Read operation type

			idx -= 1
			if idx < 16 {
				break
			}

			if _, err := file.Seek(idx, io.SeekStart); err != nil {

				return nil, err
			}

			opTypeBuf := make([]byte, 1)

			_, err := file.Read(opTypeBuf)
			if err != nil {
				return nil, err
			}

			// Read key length and key
			idx -= int64(keyLengthSize)
			if idx < 16 {
				break
			}

			if _, err := file.Seek(idx, io.SeekStart); err != nil {
				return nil, err
			}
			keyLenBuf := make([]byte, keyLengthSize)
			_, err = file.Read(keyLenBuf)
			if err != nil {
				return nil, err
			}
			keyLen := binary.BigEndian.Uint32(keyLenBuf)
			idx -= int64(keyLen)
			if idx < 16 {
				break
			}
			if _, err := file.Seek(idx, io.SeekStart); err != nil {
				return nil, err
			}
			keyBuf := make([]byte, keyLen)
			_, err = file.Read(keyBuf)
			if err != nil {
				return nil, err
			}

			// Check if the key matches
			if bytes.Equal(key, keyBuf) {
				if opTypeBuf[0] == 1 {
					// Set operation
					idx -= int64(valueLengthSize)
					if idx < 16 {
						break
					}
					valueLenBuf := make([]byte, valueLengthSize)
					if _, err := file.Seek(idx, io.SeekStart); err != nil {
						return nil, err
					}
					_, err := file.Read(valueLenBuf)
					if err != nil {
						return nil, err
					}

					valueLen := binary.BigEndian.Uint32(valueLenBuf)
					idx -= int64(valueLen)
					if idx < 16 {
						break
					}
					valueBuf := make([]byte, valueLen)
					if _, err := file.Seek(idx, io.SeekStart); err != nil {
						return nil, err
					}
					_, err = file.Read(valueBuf)
					if err != nil {
						return nil, err
					}

					return valueBuf, nil
				} else {
					// Delete operation
					fmt.Println("Delete Operation")
					return nil, errors.New("key not found")
				}
			} else {
				// Set operation
				idx -= int64(valueLengthSize)
				if idx < 16 {
					break
				}
				valueLenBuf := make([]byte, valueLengthSize)
				if _, err := file.Seek(idx, io.SeekStart); err != nil {
					return nil, err
				}
				_, err := file.Read(valueLenBuf)
				if err != nil {
					return nil, err
				}

				valueLen := binary.BigEndian.Uint32(valueLenBuf)
				idx -= int64(valueLen)
				if idx < 16 {
					break
				}
				valueBuf := make([]byte, valueLen)
				if _, err := file.Seek(idx, io.SeekStart); err != nil {
					return nil, err
				}
				_, err = file.Read(valueBuf)
				if err != nil {
					return nil, err
				}

			}
		}
	}

	// Key not found in SST filesgetSS
	fmt.Println("Key not found in SST files")
	return nil, errors.New("key not found")
}

func (mem *memDB) appendEntriesToSST(entries map[string]string) error {
	// Append entries to the current SST file
	err := mem.file.createSST(entries)
	if err != nil {
		return err
	}
	return nil
}

func (mem *memDB) FlushMemToSSTFile() error {

	entries := mem.parseMemTableEntries()

	if err := mem.createNewSSTFile(); err != nil {
		return err
	}

	if err := mem.appendEntriesToSST(entries); err != nil {
		return err
	}

	mem.memValues = nil

	return nil
}

// Calculate the total size of the new entries
func calculateEntriesSize(entries map[string]string) int64 {
	var totalSize int64

	for key, value := range entries {
		// Key length + value length + operation type size
		totalSize += int64(len(key)) + int64(len(value)) + 1
		// Key length size + value length size
		totalSize += 8
	}

	return totalSize
}

func (mem *memDB) getFileSize() (int64, error) {
	// Get the size of the current SST file
	fileInfo, err := os.Stat(fmt.Sprintf("sst_%d.sst", mem.file.noFiles))
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}

func (mem *memDB) parseMemTableEntries() map[string]string {
	// Parse entries from the memTable
	entries := make(map[string]string)
	lines := strings.Split(string(mem.memValues), "\n")
	for _, line := range lines {
		if line != "" {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				key := parts[0] + " " + parts[1]
				switch parts[0] {
				case "set":
					if len(parts) >= 3 {
						entries[key] = parts[2]
					}
				case "del":
					// For simplicity, assume a deleted key has an empty value
					entries[key] = ""
				}
			}
		}
	}
	return entries
}

type Repl struct {
	db  *memDB
	in  io.Reader
	out io.Writer
}

func (re *Repl) parseCmd(buf []byte) (Cmd, []string, error) {
	line := string(buf)
	elements := strings.Fields(line)
	if len(elements) < 1 {
		return Unk, nil, Empty
	}

	switch elements[0] {
	case "get":
		return Get, elements[1:], nil
	case "set":
		return Set, elements[1:], nil
	case "del":
		return Del, elements[1:], nil
	case "flush":
		return Flush, nil, nil
	case "exit":
		return Ext, nil, nil
	case "init":
		return Init, nil, nil
	case "test":
		return Test, nil, nil
	default:
		return Unk, nil, nil
	}
}

func (re *Repl) Start() {
	scanner := bufio.NewScanner(re.in)

	for {
		fmt.Fprint(re.out, "> ")
		if !scanner.Scan() {
			break
		}
		buf := scanner.Bytes()
		cmd, elements, err := re.parseCmd(buf)
		if err != nil {
			fmt.Fprintf(re.out, "%s\n", err.Error())
			continue
		}
		switch cmd {
		case Get:
			if len(elements) != 1 {
				fmt.Fprintf(re.out, "Expected 1 arguments, received: %d\n", len(elements))
				continue
			}
			v, err := re.db.Get([]byte(elements[0]))
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
			fmt.Fprintln(re.out, string(v))
		case Set:
			if len(elements) != 2 {
				fmt.Printf("Expected 2 arguments, received: %d\n", len(elements))
				continue
			}
			err := re.db.Set([]byte(elements[0]), []byte(elements[1]))
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
		case Del:
			if len(elements) != 1 {
				fmt.Printf("Expected 1 arguments, received: %d\n", len(elements))
				continue
			}
			v, err := re.db.Del(elements[0])
			if err != nil {
				fmt.Fprintln(re.out, err.Error())
				continue
			}
			fmt.Fprintln(re.out, string(v))
		case Flush:
			if elements != nil {
				fmt.Fprintf(re.out, "Can only use flush alone (command : flush)")
				continue
			}

			fmt.Println("WAL flushed to disk !")
		case Init:
			fmt.Println("Init")
		case Test:
			fmt.Println("Testing !")
			re.db.FlushMemToSSTFile()
		case Ext:
			fmt.Fprintln(re.out, "Bye!")
			return
		case Unk:
			fmt.Fprintln(re.out, "Unkown command")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(re.out, err.Error())
	} else {
		fmt.Fprintln(re.out, "Bye!")
	}
}