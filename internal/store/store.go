package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type Article struct {
	ID          string `json:"id"`
	XinzhiID    string `json:"xinzhiId"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	Markdown    string `json:"markdown"`
	AuthorID    string `json:"authorId"`
	AuthorName  string `json:"authorName"`
	CreatedAt   int64  `json:"createdAt"`
	SyncedAt    int64  `json:"syncedAt"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"-"`
	FeedToken string `json:"feedToken"`
}

type SyncStatus struct {
	LastSyncAt    int64  `json:"lastSyncAt"`
	NextSyncAt    int64  `json:"nextSyncAt"`
	ArticleCount  int    `json:"articleCount"`
	LastError     string `json:"lastError"`
	IsRunning     bool   `json:"isRunning"`
}

type Store struct {
	db *sql.DB
}

func New(dbPath string) (*Store, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			xinzhi_id TEXT UNIQUE NOT NULL,
			title TEXT NOT NULL,
			link TEXT,
			description TEXT,
			markdown TEXT,
			author_id TEXT,
			author_name TEXT,
			created_at INTEGER NOT NULL,
			synced_at INTEGER NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_articles_created ON articles(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_articles_xinzhi ON articles(xinzhi_id);

		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			feed_token TEXT UNIQUE NOT NULL
		);

		CREATE TABLE IF NOT EXISTS sync_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			synced_at INTEGER NOT NULL,
			article_count INTEGER,
			error TEXT
		);

		CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		);
	`)
	return err
}

func (s *Store) EnsureAdmin(username, password string) error {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	if count > 0 {
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	token := generateToken()
	_, err = s.db.Exec("INSERT INTO users (username, password, feed_token) VALUES (?, ?, ?)",
		username, string(hash), token)
	return err
}

func (s *Store) ValidateUser(username, password string) (*User, error) {
	u := &User{}
	err := s.db.QueryRow("SELECT id, username, password, feed_token FROM users WHERE username = ?",
		username).Scan(&u.ID, &u.Username, &u.Password, &u.FeedToken)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid password")
	}
	return u, nil
}

func (s *Store) GetUserByID(id int) (*User, error) {
	u := &User{}
	err := s.db.QueryRow("SELECT id, username, password, feed_token FROM users WHERE id = ?",
		id).Scan(&u.ID, &u.Username, &u.Password, &u.FeedToken)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) ValidateFeedToken(token string) bool {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM users WHERE feed_token = ?", token).Scan(&count)
	return count > 0
}

func (s *Store) RegenerateFeedToken(userID int) (string, error) {
	token := generateToken()
	_, err := s.db.Exec("UPDATE users SET feed_token = ? WHERE id = ?", token, userID)
	return token, err
}

func (s *Store) UpsertArticle(a *Article) error {
	_, err := s.db.Exec(`
		INSERT INTO articles (xinzhi_id, title, link, description, markdown, author_id, author_name, created_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(xinzhi_id) DO UPDATE SET
			title = excluded.title,
			link = excluded.link,
			description = excluded.description,
			markdown = excluded.markdown,
			synced_at = excluded.synced_at
	`, a.XinzhiID, a.Title, a.Link, a.Description, a.Markdown, a.AuthorID, a.AuthorName, a.CreatedAt, a.SyncedAt)
	return err
}

func (s *Store) ListArticles(page, pageSize int, keyword string) ([]Article, int, error) {
	offset := (page - 1) * pageSize
	var total int
	var rows *sql.Rows
	var err error

	if keyword != "" {
		like := "%" + keyword + "%"
		s.db.QueryRow("SELECT COUNT(*) FROM articles WHERE title LIKE ? OR description LIKE ?",
			like, like).Scan(&total)
		rows, err = s.db.Query(`
			SELECT id, xinzhi_id, title, link, description, '', author_id, author_name, created_at, synced_at
			FROM articles WHERE title LIKE ? OR description LIKE ?
			ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			like, like, pageSize, offset)
	} else {
		s.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&total)
		rows, err = s.db.Query(`
			SELECT id, xinzhi_id, title, link, description, '', author_id, author_name, created_at, synced_at
			FROM articles ORDER BY created_at DESC LIMIT ? OFFSET ?`,
			pageSize, offset)
	}
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.XinzhiID, &a.Title, &a.Link, &a.Description,
			&a.Markdown, &a.AuthorID, &a.AuthorName, &a.CreatedAt, &a.SyncedAt); err != nil {
			return nil, 0, err
		}
		articles = append(articles, a)
	}
	return articles, total, nil
}

func (s *Store) GetArticle(id string) (*Article, error) {
	a := &Article{}
	err := s.db.QueryRow(`
		SELECT id, xinzhi_id, title, link, description, markdown, author_id, author_name, created_at, synced_at
		FROM articles WHERE id = ?`, id).Scan(
		&a.ID, &a.XinzhiID, &a.Title, &a.Link, &a.Description,
		&a.Markdown, &a.AuthorID, &a.AuthorName, &a.CreatedAt, &a.SyncedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetAllArticlesForFeed() ([]Article, error) {
	rows, err := s.db.Query(`
		SELECT id, xinzhi_id, title, link, description, markdown, author_id, author_name, created_at, synced_at
		FROM articles ORDER BY created_at DESC LIMIT 100`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article
		if err := rows.Scan(&a.ID, &a.XinzhiID, &a.Title, &a.Link, &a.Description,
			&a.Markdown, &a.AuthorID, &a.AuthorName, &a.CreatedAt, &a.SyncedAt); err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

func (s *Store) ArticleCount() int {
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM articles").Scan(&count)
	return count
}

func (s *Store) LogSync(count int, syncErr string) {
	s.db.Exec("INSERT INTO sync_log (synced_at, article_count, error) VALUES (?, ?, ?)",
		time.Now().UnixMilli(), count, syncErr)
}

func (s *Store) GetLastSync() (int64, string) {
	var syncedAt int64
	var errStr sql.NullString
	s.db.QueryRow("SELECT synced_at, error FROM sync_log ORDER BY synced_at DESC LIMIT 1").
		Scan(&syncedAt, &errStr)
	if errStr.Valid {
		return syncedAt, errStr.String
	}
	return syncedAt, ""
}

func (s *Store) Close() error {
	return s.db.Close()
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
