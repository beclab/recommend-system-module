package service

import (
	"fmt"
	"io/fs"
	"math"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"bytetrade.io/web3os/backend-server/common"
	"bytetrade.io/web3os/backend-server/storage"
	"bytetrade.io/web3os/fs-lib/jfsnotify"

	"github.com/rs/zerolog/log"
)

var watcher *jfsnotify.Watcher = nil

func WatchPath(store *storage.Storage, deletePaths []string) {
	fmt.Println("Begin watching path...", deletePaths)
	var err error
	if watcher == nil {
		watcher, err = jfsnotify.NewWatcher("filesWatcher")
		if err != nil {
			log.Error().Msgf("new watch error %s", err.Error())
			return
		}
		go dedupLoop(store, watcher)
		log.Info().Msgf("watching path %s", strings.Join(deletePaths, ","))
	}
	for _, path := range deletePaths {
		err = filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				fmt.Println("add path...", info.Name())
				err = watcher.Add(path)
				if err != nil {
					fmt.Println("watcher add error:", err)
					return err
				}
			} else {
			}
			return nil
		})
		if err != nil {
			log.Error().Msgf("new watch error %s", err.Error())
		}
	}
}

func dedupLoop(store *storage.Storage, w *jfsnotify.Watcher) {
	var (
		// Wait 1000ms for new events; each new event resets the timer.
		waitFor = 1000 * time.Millisecond

		// Keep track of the timers, as path → timer.
		mu           sync.Mutex
		timers       = make(map[string]*time.Timer)
		pendingEvent = make(map[string]jfsnotify.Event)

		// Callback we run.
		printEvent = func(e jfsnotify.Event) {
			log.Info().Msgf("handle event %v %v", e.Op.String(), e.Name)

			// Don't need to remove the timer if you don't have a lot of files.
			mu.Lock()
			delete(pendingEvent, e.Name)
			delete(timers, e.Name)
			mu.Unlock()
		}
	)

	for {
		select {
		// Read from Errors.
		case err, ok := <-w.Errors:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				return
			}
			printTime("ERROR: %s", err)
		// Read from Events.
		case e, ok := <-w.Events:
			if !ok { // Channel was closed (i.e. Watcher.Close() was called).
				log.Warn().Msg("watcher event channel closed")
				return
			}
			if e.Has(jfsnotify.Chmod) {
				continue
			}
			//log.Debug().Msgf("pending event %v", e)
			// Get timer.
			mu.Lock()
			pendingEvent[e.Name] = e
			t, ok := timers[e.Name]
			mu.Unlock()

			// No timer yet, so create one.
			if !ok {
				t = time.AfterFunc(math.MaxInt64, func() {
					mu.Lock()
					ev := pendingEvent[e.Name]
					mu.Unlock()
					printEvent(ev)
					err := handleEvent(store, ev)
					if err != nil {
						log.Error().Msgf("handle watch file event error %s", err.Error())
					}
				})
				t.Stop()

				mu.Lock()
				timers[e.Name] = t
				mu.Unlock()
			}

			t.Reset(waitFor)
		}
	}
}

func handleEvent(store *storage.Storage, e jfsnotify.Event) error {
	if e.Has(jfsnotify.Remove) || e.Has(jfsnotify.Rename) {
		log.Info().Msgf("push indexer task delete %s", e.Name)
		//fileName := filepath.Base(e.Name)
		fileName := strings.Replace(e.Name, common.GetWatchDir()+"/", "", 1)
		entries := store.GetEntryByLocalFileName(fileName)
		log.Info().Msgf("file match entry num filename %s, %d", fileName, len(entries))
		if len(entries) > 0 {
			for _, entry := range entries {
				store.UpdateEntryFileRemove(entry.ID)
			}
		} else {
			enclosureIDs := store.GetEnclosureByLocalFileName(fileName)
			log.Info().Msgf("file match enclosure num %d", len(enclosureIDs))
			for _, enclosureID := range enclosureIDs {
				store.UpdateEnclosureFileRemove(enclosureID)
			}
		}

	}

	if e.Has(jfsnotify.Create) { // || e.Has(jfsnotify.Write) || e.Has(jfsnotify.Chmod) {
		err := filepath.Walk(e.Name, func(docPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				err = watcher.Add(docPath)
				if err != nil {
					log.Error().Msgf("watcher add error:%v", err)
				}
			}
			return nil
		})
		if err != nil {
			log.Error().Msgf("handle create file error %v", err)
		}
		return nil
	}

	if e.Has(jfsnotify.Write) {
		//return updateOrInputDoc(e.Name)
	}
	return nil
}

func printTime(s string, args ...interface{}) {
	log.Info().Msgf(time.Now().Format("15:04:05.0000")+" "+s+"\n", args...)
}
