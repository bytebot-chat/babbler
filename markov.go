package main

import (
	"strings"
	"sync"

	"github.com/mb-14/gomarkov"
	"github.com/rs/zerolog/log"
)

type model struct {
	chain *gomarkov.Chain
	mutex *sync.RWMutex
}

func (m *model) train(text string) {
	m.mutex.Lock()
	m.chain.Add(strings.Split(text, " "))
	log.Debug().RawJSON("text", []byte(text)).Msg("Model updated")
	m.mutex.Unlock()
}

func (m *model) babble() string {
	tokens := []string{gomarkov.StartToken}
	length := 0
	for tokens[len(tokens)-1] != gomarkov.EndToken {
		next, _ := m.chain.Generate(tokens[(len(tokens) - 1):])
		length += len(next)
		if length > 200 {
			tokens = append(tokens, gomarkov.EndToken)
		} else {
			tokens = append(tokens, next)
		}
	}
	return strings.Join(tokens[1:len(tokens)-1], " ")
}
