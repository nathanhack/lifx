package internal

import "sync"

var headerSequence byte
var headerSequenceMux sync.Mutex

func GetNextSequence() byte {
	headerSequenceMux.Lock()
	defer headerSequenceMux.Unlock()

	headerSequence++
	return headerSequence
}
