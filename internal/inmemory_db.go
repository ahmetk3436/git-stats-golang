package internal

import (
	"fmt"
	"sync"
)

type CommitStats struct {
	Additions int
	Deletions int
	Total     int
}

type InMemoryCommitDB struct {
	mu          sync.Mutex
	commitStats map[string]CommitStats
}

func NewInMemoryCommitDB() *InMemoryCommitDB {
	return &InMemoryCommitDB{
		commitStats: make(map[string]CommitStats),
	}
}

func (db *InMemoryCommitDB) AddCommitStats(author string, stats CommitStats) {
	db.mu.Lock()
	defer db.mu.Unlock()

	existingStats, ok := db.commitStats[author]
	if ok {
		existingStats.Additions += stats.Additions
		existingStats.Deletions += stats.Deletions
		existingStats.Total += stats.Total
		db.commitStats[author] = existingStats
	} else {
		db.commitStats[author] = stats
	}
}

func (db *InMemoryCommitDB) GetCommitStats(author string) (CommitStats, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	stats, ok := db.commitStats[author]
	if !ok {
		return CommitStats{}, fmt.Errorf("Yazar bulunamadı: %s", author)
	}
	return stats, nil
}

func (db *InMemoryCommitDB) GetAllCommitStats() map[string]CommitStats {
	db.mu.Lock()
	defer db.mu.Unlock()

	allStats := make(map[string]CommitStats)
	for author, stats := range db.commitStats {
		allStats[author] = stats
	}
	return allStats
}
