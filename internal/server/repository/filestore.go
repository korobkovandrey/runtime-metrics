package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
)

func (s Store) pack() ([]byte, error) {
	data, err := json.MarshalIndent(s.GetAllData(), "", "   ")
	if err != nil {
		return nil, fmt.Errorf("pack: %w", err)
	}
	return data, nil
}

func (s Store) unpack(data []byte) error {
	var r []StorageValue
	err := json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("UnPack: %w", err)
	}
	for _, v := range r {
		i, _ := v.Value.(float64)
		switch v.T {
		case gaugeType:
			s.repository.SetFloat64(v.T, v.Name, i)
		case counterType:
			s.repository.SetInt64(v.T, v.Name, int64(i))
		}
	}
	return nil
}

func (s Store) restore() error {
	if s.config.FileStoragePath == "" {
		return nil
	}
	stat, err := os.Stat(s.config.FileStoragePath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed restore: %w", err)
		}
		return nil
	}
	if stat.Size() > 0 {
		data, err := os.ReadFile(s.config.FileStoragePath)
		if err != nil {
			return fmt.Errorf("failed restore: %w", err)
		}
		if len(data) > 0 {
			err = s.unpack(data)
			if err != nil {
				return fmt.Errorf("failed restore: %w", err)
			}
		}
	}
	return nil
}

func (s Store) save() error {
	if s.config.FileStoragePath == "" {
		return nil
	}
	data, err := s.pack()
	if err != nil {
		return fmt.Errorf("failed save: %w", err)
	}
	const (
		permFlag = 0o600
	)
	err = os.WriteFile(s.config.FileStoragePath, data, permFlag)
	if err != nil {
		return fmt.Errorf("failed save: %w", err)
	}
	return nil
}

func (s Store) SyncSave() error {
	if s.config.StoreInterval > 0 {
		return nil
	}
	return s.save()
}

func (s Store) Run() {
	if s.config.StoreInterval <= 0 {
		return
	}
	var err error
	for range time.NewTicker(time.Duration(s.config.StoreInterval) * time.Second).C {
		if err = s.save(); err != nil {
			log.Println(err)
		}
	}
}
