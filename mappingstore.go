package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/fsnotify.v1"
)

type mappingStore struct {
	path    string
	data    map[string]string
	watcher *fsnotify.Watcher
	lock    *sync.RWMutex
}

func (m *mappingStore) updateStore() error {
	m.lock.Lock()
	defer m.lock.Unlock()
	fp, err := os.Open(m.path)
	if err != nil {
		return err
	}
	defer fp.Close()
	r := csv.NewReader(fp)
	for {
		record, err := r.Read()
		if err == nil {
			m.data[record[0]] = record[1]
		} else if err == io.EOF {
			break
		} else {
			return err
		}
	}
	log.Print("Store updated")
	return nil
}

func (m *mappingStore) handleEvent(evt fsnotify.Event) error {
	if evt.Op == fsnotify.Rename {
		log.Printf("File was renamed.")
		m.handleRename(evt)
	}
	if evt.Op == fsnotify.Remove {
		log.Printf("File was removed. Continuing to use old data.")
		return nil
	}
	return m.updateStore()
}

func (m *mappingStore) handleRename(evt fsnotify.Event) {
	m.lock.Lock()
	log.Printf("New file name: %s", evt.Name)
	m.path = evt.Name
	defer m.lock.Unlock()
}

func (m *mappingStore) LookupPlate(plate string) (string, bool) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	result, ok := m.data[plate]
	return result, ok

}

func (m *mappingStore) Start() error {
	if err := m.updateStore(); err != nil {
		return err
	}
	go func() {
		for {
			select {
			case evt := <-m.watcher.Events:
				m.handleEvent(evt)
			case err := <-m.watcher.Errors:
				fmt.Println(err.Error())
			}
		}
	}()
	return m.watcher.Add(m.path)
}

func (m *mappingStore) Stop() error {
	return m.watcher.Close()
}

func newMappingStore(path string) (*mappingStore, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &mappingStore{
		path:    path,
		watcher: watcher,
		lock:    &sync.RWMutex{},
		data:    make(map[string]string),
	}, err
}
