package sync

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/store"
	"github.com/XimilalaXiang/ReferenceAnswerRSS/internal/xinzhi"
)

type Service struct {
	client     *xinzhi.Client
	store      *store.Store
	authorID   string
	authorName string
	interval   time.Duration
	running    atomic.Bool
	lastSync   int64
	lastErr    string
	stopCh     chan struct{}
}

func New(client *xinzhi.Client, st *store.Store, authorID string, interval time.Duration) *Service {
	return &Service{
		client:     client,
		store:      st,
		authorID:   authorID,
		authorName: "参考答案阅览室",
		interval:   interval,
		stopCh:     make(chan struct{}),
	}
}

func (s *Service) Start() {
	go s.loop()
	log.Printf("[sync] started, interval=%v, author=%s", s.interval, s.authorID)
}

func (s *Service) Stop() {
	close(s.stopCh)
}

func (s *Service) loop() {
	s.RunOnce()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.RunOnce()
		case <-s.stopCh:
			log.Println("[sync] stopped")
			return
		}
	}
}

func (s *Service) RunOnce() {
	if !s.running.CompareAndSwap(false, true) {
		log.Println("[sync] already running, skip")
		return
	}
	defer s.running.Store(false)

	log.Println("[sync] starting sync...")
	start := time.Now()
	count, err := s.syncAll()

	s.lastSync = time.Now().UnixMilli()
	if err != nil {
		s.lastErr = err.Error()
		s.store.LogSync(count, err.Error())
		log.Printf("[sync] completed with error: %v (synced %d articles in %v)", err, count, time.Since(start))
	} else {
		s.lastErr = ""
		s.store.LogSync(count, "")
		log.Printf("[sync] completed: synced %d articles in %v", count, time.Since(start))
	}
}

func (s *Service) syncAll() (int, error) {
	totalSynced := 0
	pageIndex := 1
	pageSize := 50

	for {
		resp, err := s.client.ListNotesByAuthor(s.authorID, pageIndex, pageSize)
		if err != nil {
			return totalSynced, err
		}

		for _, note := range resp.List {
			detail, err := s.client.GetNote(note.ID)
			if err != nil {
				log.Printf("[sync] failed to get note %s: %v", note.ID, err)
				continue
			}

			article := &store.Article{
				XinzhiID:    detail.ID,
				Title:       detail.Title,
				Link:        detail.Link,
				Description: detail.Description,
				Markdown:    detail.Markdown,
				AuthorID:    s.authorID,
				AuthorName:  s.authorName,
				CreatedAt:   detail.CreateTime,
				SyncedAt:    time.Now().UnixMilli(),
			}

			if err := s.store.UpsertArticle(article); err != nil {
				log.Printf("[sync] failed to upsert article %s: %v", detail.ID, err)
				continue
			}
			totalSynced++
		}

		if !resp.HasMore {
			break
		}
		pageIndex++

		time.Sleep(500 * time.Millisecond)
	}

	return totalSynced, nil
}

func (s *Service) Status() store.SyncStatus {
	return store.SyncStatus{
		LastSyncAt:   s.lastSync,
		NextSyncAt:   s.lastSync + s.interval.Milliseconds(),
		ArticleCount: s.store.ArticleCount(),
		LastError:    s.lastErr,
		IsRunning:    s.running.Load(),
	}
}
