package history

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/karimsa/patrol/internal/logger"
)

type Item struct {
	ID         string
	Group      string
	Name       string
	Type       string
	Output     []byte
	CreatedAt  time.Time
	Duration   time.Duration
	Metric     float64
	MetricUnit string
	Status     string
	Error      string
}

func (item Item) String() string {
	return strings.Join([]string{
		fmt.Sprintf("Item{"),
		fmt.Sprintf("\tGroup: %s,", item.Group),
		fmt.Sprintf("\tName: %s,", item.Name),
		fmt.Sprintf("\tType: %s,", item.Type),
		fmt.Sprintf("\tOutput: '%s',", strings.Join(strings.Split(string(item.Output), "\n"), "\\n")),
		fmt.Sprintf("\tCreatedAt: %s,", item.CreatedAt),
		fmt.Sprintf("\tDuration: %s,", item.Duration),
		fmt.Sprintf("\tMetric: %.2f %s,", item.Metric, item.MetricUnit),
		fmt.Sprintf("\tStatus: %s,", item.Status),
		fmt.Sprintf("\tError: '%s',", item.Error),
		fmt.Sprintf("}"),
	}, "\n")
}

func (item Item) writeTo(out io.Writer) error {
	data, err := json.Marshal(item)
	if err != nil {
		return err
	}
	n, err := out.Write(append(data, '\n'))
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

type listNode struct {
	value Item
	next  *listNode
	prev  *listNode
}

func (ln *listNode) String() string {
	return ln.value.String()
}

type dataContainer struct {
	byID map[string]*listNode
	head *listNode
	tail *listNode
}

type File struct {
	fd          *os.File
	writes      chan writeRequest
	writerWg    *sync.WaitGroup
	done        chan bool
	data        map[string]map[string]*dataContainer
	validGroups map[string]map[string]bool
	rwMux       *sync.RWMutex
	maxEntries  int
	logger      logger.Logger
}

type NewOptions struct {
	File                string
	MaxEntries          int
	MaxConcurrentWrites int
	Groups              map[string]map[string]bool
	LogLevel            logger.LogLevel
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
		fd:          fd,
		writes:      make(chan writeRequest, options.MaxConcurrentWrites),
		writerWg:    &sync.WaitGroup{},
		done:        make(chan bool),
		data:        map[string]map[string]*dataContainer{},
		validGroups: options.Groups,
		rwMux:       &sync.RWMutex{},
		maxEntries:  options.MaxEntries,
	}
	if file.validGroups == nil {
		file.validGroups = make(map[string]map[string]bool)
	}
	file.SetLogLevel(options.LogLevel)
	file.logger.Debugf("Opened history file: %s", options.File)

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

	file.writerWg.Add(1)
	go file.bgWriter()
	return file, nil
}

func (file *File) Compact() (numItems int, err error) {
	file.rwMux.Lock()
	defer file.rwMux.Unlock()

	writeBuffer := &bytes.Buffer{}
	for _, group := range file.data {
		for _, container := range group {
			for curr := container.head; curr != nil; curr = curr.next {
				item := curr.value
				if checkerNames, ok := file.validGroups[item.Group]; !ok {
					if _, ok := checkerNames[item.Name]; !ok {
						err = curr.value.writeTo(writeBuffer)
						if err != nil {
							return
						}
						numItems += 1
					}
				}
			}
		}
	}

	err = file.fd.Truncate(0)
	if err != nil {
		return
	}

	_, err = file.fd.Seek(0, 0)
	if err != nil {
		return
	}

	_, err = io.Copy(file.fd, writeBuffer)
	if err != nil {
		return
	}

	err = file.fd.Sync()
	if err != nil {
		return
	}

	file.logger.Infof("Data compacted - %d groups and %d items in history", len(file.data), numItems)
	return
}

func (file *File) SetLogLevel(level logger.LogLevel) {
	file.logger = logger.New(level, "history:")
}

type checker interface {
	GetGroup() string
	GetName() string
}

func (file *File) AddChecker(c checker) {
	file.rwMux.RLock()
	checkers, ok := file.validGroups[c.GetGroup()]
	if !ok {
		checkers = make(map[string]bool, 1)
		file.validGroups[c.GetGroup()] = checkers
	}
	checkers[c.GetName()] = true
	file.rwMux.RUnlock()
}

func (file *File) bgWriter() {
	defer file.writerWg.Done()

	for {
		select {
		case req := <-file.writes:
			file.rwMux.Lock()
			records := make([]writeRequest, 1)
			records[0] = req

			// TODO: Instead of panicing, it should perform a
			// rollback on the 'data' state

			buffer := &bytes.Buffer{}
			req.item = file.addItem(req.item)
			if err := req.item.writeTo(buffer); err != nil {
				sendError(records, err)
			} else {
				collect := true
				var err error

				for collect && err != nil {
					select {
					case r := <-file.writes:
						records = append(records, r)
						r.item = file.addItem(r.item)
						err = r.item.writeTo(buffer)
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
				} else {
					file.logger.Debugf("Wrote %d records", len(records))
					file.rwMux.Unlock()
					sendError(records, nil)
				}

				if err := file.fd.Sync(); err != nil {
					file.logger.Warnf("fsync failed: %s", err)
				}
			}

		case <-file.done:
			file.logger.Debugf("Closing history file")
			return
		}
	}
}

func (file *File) addItem(item Item) Item {
	if _, ok := file.data[item.Group]; !ok {
		file.data[item.Group] = make(map[string]*dataContainer, 1)
	}
	if _, ok := file.data[item.Group][item.Name]; !ok {
		file.data[item.Group][item.Name] = &dataContainer{
			byID: make(map[string]*listNode, 100),
		}
	}
	container := file.data[item.Group][item.Name]

	if item.Type == "boolean" {
		item.ID = fmt.Sprintf("%s|%s|%d|0", item.Group, item.Name, item.CreatedAt.UTC().UnixNano()/int64(24*time.Hour))
	} else {
		n := int64(0)
		prefix := fmt.Sprintf("%s|%s|%d|", item.Group, item.Name, item.CreatedAt.UTC().UnixNano())
		for {
			item.ID = prefix + strconv.FormatInt(n, 10)
			if _, exists := container.byID[item.ID]; !exists {
				break
			}
			n++
		}
	}

	node, exists := container.byID[item.ID]
	if !exists {
		node = &listNode{}
		container.byID[item.ID] = node
	}

	lastValue := node.value
	if item.Type == "boolean" && item.Status == "healthy" && (lastValue.Status == "unhealthy" || lastValue.Status == "recovered") {
		item.Status = "recovered"
	}

	// Writes to "item" after this will have no effect
	node.value = item

	if item.Type == "metric" || !exists {
		file.logger.Debugf("Inserting (size = %d): %s", len(container.byID), item)

		if container.head == nil {
			container.head = node
			container.tail = node
		} else {
			inserted := false
			var prev *listNode

			for curr := container.head; curr != nil && !inserted; prev, curr = curr, curr.next {
				if !node.value.CreatedAt.Before(curr.value.CreatedAt) {
					if prev == nil {
						node.next = container.head
						container.head.prev = node
						container.head = node
					} else {
						node.prev = prev
						node.next = prev.next
						prev.next = node
					}
					inserted = true
				}
			}

			if !inserted {
				container.tail.next = node
				node.prev = container.tail
				container.tail = node
			}

			for len(container.byID) > file.maxEntries {
				drop := container.tail
				file.logger.Debugf("Dropping old item: %s", drop.value)
				container.tail = drop.prev
				if container.tail == nil {
					container.head = nil
				} else {
					container.tail.next = nil
				}
				delete(container.byID, drop.value.ID)
			}
		}
	} else {
		file.logger.Debugf("Replacing: %s", item)
	}

	return item
}

func (file *File) Append(item Item) error {
	errChan := make(chan error)
	item.CreatedAt = time.Now()
	file.writes <- writeRequest{
		item:    item,
		errChan: errChan,
	}
	return <-errChan
}

func (file *File) GetData() map[string]map[string][]Item {
	data := make(map[string]map[string][]Item, len(file.data))
	file.rwMux.RLock()

	for groupName, group := range file.data {
		data[groupName] = make(map[string][]Item)
		for checkName, container := range group {
			list := make([]Item, 0, len(container.byID))
			for curr := container.head; curr != nil; curr = curr.next {
				list = append(list, curr.value)
			}
			data[groupName][checkName] = list
		}
	}

	file.rwMux.RUnlock()
	return data
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

func (file *File) GetGroupItems(group, checkName string) []Item {
	file.rwMux.RLock()
	g, _ := file.data[group]
	container, _ := g[checkName]
	file.rwMux.RUnlock()

	list := make([]Item, 0, len(container.byID))
	for curr := container.head; curr != nil; curr = curr.next {
		list = append(list, curr.value)
	}
	return list
}

func (file *File) Close() {
	close(file.done)
	file.writerWg.Wait()
}
