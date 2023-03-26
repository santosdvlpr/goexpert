package repository

import (
	"database/sql"
	"errors"
	"log"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrDuplicate    = errors.New("record already exists")
	ErrNotExists    = errors.New("row not exists")
	ErrUpdateFailed = errors.New("update failed")
	ErrDeleteFailed = errors.New("delete failed")
)

type SQLiteRepository struct{ db *sql.DB }

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{
		db: db,
	}
}
func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS cotacao(
		id INTEGER PRIMARY KEY,
		valor decimal(10,2)
	);
	`
	_, err := r.db.Exec(query)
	return err
}

func (r *SQLiteRepository) Create(cotacao Cotacao) (*Cotacao, error) {

	log.Print("Valor gravado:", cotacao.Valor)

	res, err := r.db.Exec("INSERT INTO cotacao(valor) values(?)", cotacao.Valor)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			if errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
				return nil, ErrDuplicate
			}
		}
		return nil, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	cotacao.ID = id

	return &cotacao, nil
}

func (r *SQLiteRepository) All() ([]Cotacao, error) {
	rows, err := r.db.Query("SELECT * FROM cotacao")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []Cotacao
	for rows.Next() {
		var cotacao Cotacao
		if err := rows.Scan(&cotacao.ID, &cotacao.Valor); err != nil {
			return nil, err
		}
		all = append(all, cotacao)
	}
	return all, nil
}

func (r *SQLiteRepository) GetById(id int64) (*Cotacao, error) {
	row := r.db.QueryRow("SELECT * FROM cotacao WHERE id = ?", id)

	var cotacao Cotacao
	if err := row.Scan(&cotacao.ID, &cotacao.Valor); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotExists
		}
		return nil, err
	}
	return &cotacao, nil
}

func (r *SQLiteRepository) Update(id int64, updated Cotacao) (*Cotacao, error) {
	if id == 0 {
		return nil, errors.New("invalid updated ID")
	}
	res, err := r.db.Exec("UPDATE cotacao SET valor = ? WHERE id = ?", updated.Valor, id)
	if err != nil {
		return nil, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrUpdateFailed
	}

	return &updated, nil
}

func (r *SQLiteRepository) Delete(id int64) error {
	res, err := r.db.Exec("DELETE FROM cotacao WHERE id = ?", id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrDeleteFailed
	}

	return err
}
