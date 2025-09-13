package recommendation

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
)

type SQLiteStorage struct {
	db     *sql.DB
	logger *logrus.Logger
	path   string
}

func NewSQLiteStorage(dbPath string, logger *logrus.Logger) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{
		db:     db,
		logger: logger,
		path:   dbPath,
	}

	if err := storage.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	logger.Info("SQLite storage initialized successfully")
	return storage, nil
}

func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			user_id INTEGER PRIMARY KEY,
			age INTEGER,
			gender TEXT,
			occupation TEXT,
			zip_code TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS movies (
			movie_id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			genres TEXT,
			year INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS ratings (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			movie_id INTEGER NOT NULL,
			rating REAL NOT NULL,
			timestamp INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (user_id) REFERENCES users(user_id),
			FOREIGN KEY (movie_id) REFERENCES movies(movie_id),
			UNIQUE(user_id, movie_id)
		)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	return nil
}

func (s *SQLiteStorage) LoadMovieLensData(dataset string) error {
	s.logger.Infof("Loading MovieLens %s dataset", dataset)
	
	// Determine data path based on dataset
	var dataPath string
	switch dataset {
	case "100k":
		dataPath = "../data/movielens/ml-100k"
		// Try different relative paths
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			dataPath = "../../data/movielens/ml-100k"
		}
	case "1m":
		dataPath = "../data/movielens/ml-1m"
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			dataPath = "../../data/movielens/ml-1m"
		}
	case "25m":
		dataPath = "../data/movielens/ml-25m"
		if _, err := os.Stat(dataPath); os.IsNotExist(err) {
			dataPath = "../../data/movielens/ml-25m"
		}
	default:
		return fmt.Errorf("unsupported dataset: %s", dataset)
	}
	
	// Load movies first
	if err := s.loadMovies(dataPath, dataset); err != nil {
		return fmt.Errorf("failed to load movies: %w", err)
	}
	
	// Load ratings 
	if err := s.loadRatings(dataPath, dataset); err != nil {
		return fmt.Errorf("failed to load ratings: %w", err)
	}
	
	s.logger.Info("MovieLens data loaded successfully")
	return nil
}

func (s *SQLiteStorage) GetStorageStats() *StorageStats {
	var userCount, movieCount, ratingCount int64
	
	s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	s.db.QueryRow("SELECT COUNT(*) FROM movies").Scan(&movieCount)
	s.db.QueryRow("SELECT COUNT(*) FROM ratings").Scan(&ratingCount)
	
	return &StorageStats{
		UserCount:   userCount,
		MovieCount:  movieCount,
		RatingCount: ratingCount,
		LastUpdated: time.Now(),
	}
}

func (s *SQLiteStorage) loadMovies(dataPath, dataset string) error {
	var filename string
	switch dataset {
	case "100k":
		filename = "u.item"
	case "1m":
		filename = "movies.dat"
	case "25m":
		filename = "movies.csv"
	default:
		return fmt.Errorf("unsupported dataset: %s", dataset)
	}

	filePath := filepath.Join(dataPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	s.logger.Infof("Loading movies from %s", filePath)
	
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT OR IGNORE INTO movies (movie_id, title, genres, year) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var movieID int
		var title, genres string
		var year int

		switch dataset {
		case "100k":
			// Format: movieID|title|release_date|video_release_date|imdb_url|genre_cols...
			parts := strings.Split(line, "|")
			if len(parts) < 5 {
				continue
			}
			movieID, _ = strconv.Atoi(parts[0])
			title = parts[1]
			// Extract year from title
			if strings.Contains(title, "(") && strings.Contains(title, ")") {
				start := strings.LastIndex(title, "(")
				end := strings.LastIndex(title, ")")
				if end > start {
					yearStr := title[start+1 : end]
					year, _ = strconv.Atoi(yearStr)
				}
			}
			genres = "Unknown" // Genre info is in binary format, simplified for now
		}

		if movieID > 0 && title != "" {
			_, err := stmt.Exec(movieID, title, genres, year)
			if err != nil {
				s.logger.Warnf("Failed to insert movie %d: %v", movieID, err)
				continue
			}
			count++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Infof("Loaded %d movies", count)
	return nil
}

func (s *SQLiteStorage) loadRatings(dataPath, dataset string) error {
	var filename string
	switch dataset {
	case "100k":
		filename = "u.data"
	case "1m":
		filename = "ratings.dat"
	case "25m":
		filename = "ratings.csv"
	default:
		return fmt.Errorf("unsupported dataset: %s", dataset)
	}

	filePath := filepath.Join(dataPath, filename)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, err)
	}
	defer file.Close()

	s.logger.Infof("Loading ratings from %s", filePath)
	
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert users on the fly
	userStmt, err := tx.Prepare("INSERT OR IGNORE INTO users (user_id) VALUES (?)")
	if err != nil {
		return fmt.Errorf("failed to prepare user statement: %w", err)
	}
	defer userStmt.Close()

	ratingStmt, err := tx.Prepare("INSERT OR IGNORE INTO ratings (user_id, movie_id, rating, timestamp) VALUES (?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare rating statement: %w", err)
	}
	defer ratingStmt.Close()

	scanner := bufio.NewScanner(file)
	count := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var userID, movieID, timestamp int
		var rating float64

		switch dataset {
		case "100k":
			// Format: userID movieID rating timestamp (tab-separated)
			parts := strings.Split(line, "\t")
			if len(parts) != 4 {
				continue
			}
			userID, _ = strconv.Atoi(parts[0])
			movieID, _ = strconv.Atoi(parts[1])
			rating, _ = strconv.ParseFloat(parts[2], 64)
			timestamp, _ = strconv.Atoi(parts[3])
		}

		if userID > 0 && movieID > 0 {
			// Insert user if not exists
			userStmt.Exec(userID)
			
			// Insert rating
			_, err := ratingStmt.Exec(userID, movieID, rating, timestamp)
			if err != nil {
				s.logger.Warnf("Failed to insert rating %d-%d: %v", userID, movieID, err)
				continue
			}
			count++
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	s.logger.Infof("Loaded %d ratings", count)
	return nil
}

func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}