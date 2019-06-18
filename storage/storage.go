package storage

import (
	"fmt"
)

type Storage struct {
	path          string
	numPartitions int32
	partitionID   int32
	partitions    []*Partition
}

func NewStorage(path string, partitionID, numPartitions, segLen int32) (*Storage, error) {
	var err error
	s := &Storage{
		path:          path,
		partitionID:   partitionID,
		numPartitions: numPartitions,
	}
	s.partitions = make([]*Partition, numPartitions)
	for i := int32(0); i < numPartitions; i++ {
		s.partitions[i], err = NewPartition(path, segLen)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func (s *Storage) Write(record string) (int64, error) {
	lsn, err := s.WritePartition(s.partitionID, record)
	return lsn, err
}

func (s *Storage) WritePartition(id int32, record string) (int64, error) {
	lsn, err := s.partitions[id].Write(record)
	return lsn, err
}

func (s *Storage) Assign(partitionID int32, lsn int64, length int32, gsn int64) error {
	// TODO handle errors: keep retrying
	return s.partitions[partitionID].Assign(lsn, length, gsn)
}

func (s *Storage) Read(gsn int64) (string, error) {
	return s.ReadGSN(gsn)
}

func (s *Storage) ReadGSN(gsn int64) (string, error) {
	// read my own partition first
	r, err := s.partitions[s.partitionID].ReadGSN(gsn)
	if err == nil {
		return r, nil
	}
	// if not in my own partition, check others
	for i := int32(0); i < s.numPartitions; i++ {
		if i == s.partitionID {
			continue
		}
		r, err = s.partitions[i].ReadGSN(gsn)
		if err == nil {
			return r, nil
		}
	}
	return "", fmt.Errorf("Record not found as gsn=%v", gsn)
}

func (s *Storage) ReadLSN(partitionID int32, lsn int64) (string, error) {
	return s.partitions[partitionID].ReadLSN(lsn)
}
