package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

/**
 * A simple file-based memory to keep track of sent messages with
 * the help of filesystem. Remember the ID, check if the file exists
 * in the memory directory, if it exists, skip the message. If it does
 * not exist, create the file to remember the message.
 *
 * @param mutex		A mutex to protect the memory directory.
 * @param data		The message data to be checked and remembered.
 * @return		true if the message should be sent
 *			false if it should be skipped.
 */
func IsNewMessage(m *sync.Mutex, id string) bool {
	/*
	 * Always send the message if the memory directory is not set.
	 */
	if len(config.MsgMemoryDir) == 0 {
		return true
	}

	fpath := fmt.Sprintf("%s/%x", config.MsgMemoryDir, id)

	m.Lock()
	defer m.Unlock()
	/*
	 * If the file exists, skip the message. We've already sent it.
	 */
	if _, err := os.Stat(fpath); err == nil {
		log.Printf("wrs: Skipping message, already sent: %x", id)
		return false
	}

	if err := os.MkdirAll(config.MsgMemoryDir, 0755); err != nil && !os.IsExist(err) {
		log.Printf("Failed to create message memory directory: %s", err)
		return true
	}

	/*
	 * Create the file to remember the message.
	 */
	f, err := os.Create(fpath)
	if err != nil {
		log.Printf("Failed to create message memory file: %s", err)
		return true
	}
	f.Close()

	return true
}

func ScanAndDeleteOldMessages(m *sync.Mutex) {
	m.Lock()
	defer m.Unlock()
	files, err := os.ReadDir(config.MsgMemoryDir)
	if err != nil {
		log.Printf("Failed to read message memory directory: %s", err)
		return
	}

	for _, file := range files {
		st, err := file.Info()
		if err != nil {
			log.Printf("Failed to get file info: %s", err)
			continue
		}

		if time.Since(st.ModTime()) <= 7*24*time.Hour {
			continue
		}

		fpath := fmt.Sprintf("%s/%s", config.MsgMemoryDir, file.Name())
		if err := os.Remove(fpath); err != nil {
			log.Printf("Failed to delete old message memory file: %s", err)
		} else {
			log.Printf("Deleted old message memory file: %s", fpath)
		}
	}
}

func MemDirHouseKeeping(m *sync.Mutex) {
	for {
		log.Println("wrs: Starting message memdir housekeeping")
		ScanAndDeleteOldMessages(m)
		log.Println("wrs: Finished message memdir housekeeping, will be back in 3 hours")
		time.Sleep(3 * time.Hour)
	}
}
