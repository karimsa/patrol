package history

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
	"sort"
	"bufio"
)

type Item struct {
	Group     string
	Name      string
	Type      string
	Output    []byte
	CreatedAt time.Time
}

func (item Item) writeTo(buffer *bytes.Buffer) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	n, err := buffer.Write(append(data, '\n'))
	if n < len(data) {
		return fmt.Errorf("Wrote partial data (error: %s)", err)
	}
	return err
}

func sendError(receivers []writeRequest, err error) {
	for _, recv := range receivers {
		recv.errChan <- err
	}
}

type writeRequest struct {
	item    Item
	errChan chan error
}

type File struct {
	fd       *os.File
	writes   chan writeRequest
	writerWg *sync.WaitGroup
	done     chan bool
	data     map[string][]Item
	rwMux    *sync.RWMutex
}

type NewOptions struct {
	File                string
	MaxEntries          int
	MaxConcurrentWrites int
}

func New(options NewOptions) (*File, error) {
	fd, err := os.OpenFile(
		options.File,
		os.O_RDWR|os.O_CREATE,
		0755,
	)
	if err != nil {
		return nil, err
	}

	file := &File{
		fd:       fd,
		writes:   make(chan writeRequest, options.MaxConcurrentWrites),
		writerWg: &sync.WaitGroup{},
		done:     make(chan bool),
		data:     make(map[string][]Item),
		rwMux:    &sync.RWMutex{},
	}

	// TODO: Read log file
	bufferedReader := bufio.NewReader(fd)
	var item Item
	var line []byte
	for err != io.EOF {
		line, err = bufferedReader.ReadBytes('\n')
		if len(line) > 0 {
			if err := json.Unmarshal(line[:len(line)-1], &item); err != nil {
				return nil, err
			}
			file.addItem(item)
		}
	}

	go file.bgWriter()
	return file, nil
}

func (file *File) bgWriter() {
	file.writerWg.Add(1)
	defer file.writerWg.Done()

	for {
		select {
		case req := <-file.writes:
			file.rwMux.Lock()
			records := make([]writeRequest, 1)
			records[0] = req

			buffer := bytes.NewBuffer([]byte{})
			if err := req.item.writeTo(buffer); err != nil {
				sendError(records, err)
			} else {
				collect := true
				var err error

				for collect && err != nil {
					select {
					case r := <-file.writes:
						records = append(records, r)
						err = req.item.writeTo(buffer)
					default:
						collect = false
					}
				}

				if err != nil {
					sendError(records, err)
				} else if n, err := io.Copy(file.fd, buffer); err != nil {
					file.rwMux.Unlock()
					panic(err)
				} else if n < int64(buffer.Len()) {
					file.rwMux.Unlock()
					panic(fmt.Errorf("Wrote only %d bytes to file", n))
				} else{
					for _, r := range records {
						file.addItem(r.item)					}

					file.rwMux.Unlock()
					sendError(records, nil)
				}
			}

		case <-file.done:
			return
		}
	}
}

func (file *File) addItem(item Item) {
	if _, ok := file.data[item.Group]; !ok {
							file.data[item.Group] = make([]Item, 0, 1)
						}

						lst := append(file.data[item.Group], item)
						file.data[item.Group] = lst

						sort.SliceStable(file.data[item.Group], func(i, j int) bool {
							return lst[j].CreatedAt.Before(lst[i].CreatedAt)
						})
}

func (file *File) Append(item Item) error {
	errChan := make(chan error)
	file.writes <- writeRequest{
		item:    item,
		errChan: errChan,
	}
	return <-errChan
}

func (file *File) GetGroups() []string {
	keys := make([]string, len(file.data))
	idx := 0

	file.rwMux.RLock()
	for key, _ := range file.data {
		keys[idx] = key
		idx++
	}
	file.rwMux.RUnlock()

	return keys
}

func (file *File) GetGroupItems(group string) []Item {
	file.rwMux.RLock()
	items, groupExists := file.data[group]
	file.rwMux.RUnlock()

	if groupExists {
		return items
	}
	return []Item{}
}

func (file *File) Close() {
	file.done <- true
	file.writerWg.Wait()
}
