package main

import (
	"fmt"
	"sync"
	"log"
)

// TODO: Create this service per apps.
// In this example, this instance is singleton.
type RelayService struct {
	streams map[string]*Pubsub
	m       sync.Mutex
}

func NewRelayService() *RelayService {
	return &RelayService{
		streams: make(map[string]*Pubsub),
	}
}

func (s *RelayService) NewPubsub(key string) (*Pubsub, error) {
	log.Println("NEW PUBSUB INIT")
	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.streams[key]; ok {
		return nil, fmt.Errorf("Already published: %s", key)
	}

	pubsub := NewPubsub(s, key)

	s.streams[key] = pubsub
	log.Println("NEW PUBSUB END")

	return pubsub, nil
}

func (s *RelayService) GetPubsub(key string) (*Pubsub, error) {

	s.m.Lock()
	defer s.m.Unlock()

	pubsub, ok := s.streams[key]
	if !ok {
		return nil, fmt.Errorf("Not published: %s", key)
	}

	return pubsub, nil
}

func (s *RelayService) RemovePubsub(key string) error {
	log.Println("REMOVE PUBSUB INIT")

	s.m.Lock()
	defer s.m.Unlock()

	if _, ok := s.streams[key]; !ok {
		return fmt.Errorf("Not published: %s", key)
	}

	delete(s.streams, key)
	log.Println("REMOVE PUBSUB END")

	return nil
}
