package db

import (
	"database/sql"
	"fmt"
	"tasker/src/models"

	_ "modernc.org/sqlite"
)

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("sqlite", "tasks.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := createTables(db); err != nil {
		return nil, err
	}

	if err := migrateDatabase(db); err != nil {
		return nil, err
	}

	return &Database{db: db}, nil
}

func createTables(db *sql.DB) error {
	// Create tasks table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			done BOOLEAN DEFAULT 0,
			created_at DATETIME NOT NULL,
			due_date DATETIME,
			priority INTEGER NOT NULL,
			status INTEGER NOT NULL DEFAULT 0
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tasks table: %v", err)
	}

	// Create tags table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create tags table: %v", err)
	}

	// Create task_tags junction table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_tags (
			task_id INTEGER,
			tag_id INTEGER,
			PRIMARY KEY (task_id, tag_id),
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create task_tags table: %v", err)
	}

	return nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) SaveTask(task *models.Task) error {
	tx, err := d.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback()

	// Insert or update task
	var taskID int64
	if task.ID == 0 {
		result, err := tx.Exec(`
			INSERT INTO tasks (title, description, done, created_at, due_date, priority, status)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, task.Title, task.Description, task.Done, task.CreatedAt, task.DueDate, task.Priority, task.Status)
		if err != nil {
			return fmt.Errorf("failed to insert task: %v", err)
		}
		taskID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert ID: %v", err)
		}
		task.ID = int(taskID)
	} else {
		_, err := tx.Exec(`
			UPDATE tasks
			SET title = ?, description = ?, done = ?, due_date = ?, priority = ?, status = ?
			WHERE id = ?
		`, task.Title, task.Description, task.Done, task.DueDate, task.Priority, task.Status, task.ID)
		if err != nil {
			return fmt.Errorf("failed to update task: %v", err)
		}
		taskID = int64(task.ID)
	}

	// Clear existing tags
	_, err = tx.Exec("DELETE FROM task_tags WHERE task_id = ?", taskID)
	if err != nil {
		return fmt.Errorf("failed to clear task tags: %v", err)
	}

	// Insert new tags
	for _, tag := range task.Tags {
		// Get or create tag
		var tagID int64
		err := tx.QueryRow("SELECT id FROM tags WHERE name = ?", tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			result, err := tx.Exec("INSERT INTO tags (name) VALUES (?)", tag)
			if err != nil {
				return fmt.Errorf("failed to insert tag: %v", err)
			}
			tagID, err = result.LastInsertId()
			if err != nil {
				return fmt.Errorf("failed to get tag ID: %v", err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query tag: %v", err)
		}

		// Link tag to task
		_, err = tx.Exec("INSERT INTO task_tags (task_id, tag_id) VALUES (?, ?)", taskID, tagID)
		if err != nil {
			return fmt.Errorf("failed to link tag to task: %v", err)
		}
	}

	return tx.Commit()
}

func (d *Database) LoadTasks() ([]*models.Task, error) {
	rows, err := d.db.Query(`
		SELECT t.id, t.title, t.description, t.done, t.created_at, t.due_date, t.priority, t.status,
		       GROUP_CONCAT(tg.name) as tags
		FROM tasks t
		LEFT JOIN task_tags tt ON t.id = tt.task_id
		LEFT JOIN tags tg ON tt.tag_id = tg.id
		GROUP BY t.id
		ORDER BY t.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %v", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		var task models.Task
		var tags sql.NullString
		var dueDate sql.NullTime

		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Done,
			&task.CreatedAt,
			&dueDate,
			&task.Priority,
			&task.Status,
			&tags,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %v", err)
		}

		if dueDate.Valid {
			task.DueDate = &dueDate.Time
		}

		if tags.Valid && tags.String != "" {
			task.Tags = []string{tags.String}
		} else {
			task.Tags = []string{}
		}

		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (d *Database) DeleteTask(id int) error {
	_, err := d.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}
	return nil
}
