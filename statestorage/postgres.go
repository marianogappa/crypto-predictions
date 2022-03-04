package statestorage

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/types"
)

type PostgresDBStateStorage struct {
	db *sql.DB
}

//go:embed migrations/*.sql
var fs embed.FS

func NewPostgresDBStateStorage() (PostgresDBStateStorage, error) {
	connStr := "postgres://marianol@localhost/marianol?sslmode=disable"

	d, err := iofs.New(fs, "migrations")
	if err != nil {
		return PostgresDBStateStorage{}, err
	}
	m, err := migrate.NewWithSourceInstance("iofs", d, connStr)
	if err != nil {
		return PostgresDBStateStorage{}, err
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return PostgresDBStateStorage{}, err
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return PostgresDBStateStorage{}, err
	}
	return PostgresDBStateStorage{db: db}, nil
}

type pgUpsertManyBuilder struct {
	i           int
	columnNames []string
	values      []interface{}
	rowCount    int

	strUpdate, strCols, tableName, conflictKey string
}

func newPGUpsertManyBuilder(columnNames []string, tableName, conflictKey string) *pgUpsertManyBuilder {
	builder := pgUpsertManyBuilder{columnNames: columnNames, tableName: tableName, conflictKey: conflictKey}

	update := []string{}
	for _, colName := range columnNames {
		update = append(update, fmt.Sprintf(`%v = EXCLUDED.%v`, colName, colName))
	}
	builder.strUpdate = strings.Join(update, ",")
	builder.strCols = strings.Join(columnNames, ",")

	return &builder
}

func (b *pgUpsertManyBuilder) addRow(values ...interface{}) {
	b.values = append(b.values, values...)
	b.rowCount++
}

func (b *pgUpsertManyBuilder) build() (string, []interface{}) {
	paramCount := 1
	rows := []string{}
	for i := 0; i < b.rowCount; i++ {
		params := []string{}
		for j := 0; j < len(b.columnNames); j++ {
			params = append(params, fmt.Sprintf("$%v", paramCount))
			paramCount++
		}
		rows = append(rows, fmt.Sprintf("(%v)", strings.Join(params, ", ")))
	}
	return fmt.Sprintf(`
		INSERT INTO %v (%v)
		VALUES %v 
		ON CONFLICT (%v) DO UPDATE SET %v`,
		b.tableName,
		b.strCols,
		strings.Join(rows, ","),
		b.conflictKey,
		b.strUpdate,
	), b.values
}

func (s PostgresDBStateStorage) GetPredictions(css []types.PredictionStateValue) (map[string]types.Prediction, error) {
	// TODO filter
	rows, err := s.db.Query("SELECT uuid, blob FROM predictions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[string]types.Prediction{}
	for rows.Next() {
		var (
			clUUID, clBlob []byte
		)
		err := rows.Scan(&clUUID, &clBlob)
		if err != nil {
			log.Printf("error reading predictions fields from db, with error: %v\n", err)
		}
		var pred types.Prediction
		if pred, err = compiler.NewPredictionCompiler().Compile(clBlob); err != nil {
			log.Printf("read corrupted prediction from db, with error: %v\n", err)
			continue
		}
		pred.UUID = string(clUUID)
		result[pred.PostUrl] = pred
	}

	return result, nil
}

func (s PostgresDBStateStorage) UpsertPredictions(ps map[string]types.Prediction) error {
	if len(ps) == 0 {
		return nil
	}
	builder := newPGUpsertManyBuilder([]string{"uuid", "blob", "created_at", "posted_at"}, "predictions", "uuid")
	for _, p := range ps {
		if p.UUID == "" {
			p.UUID = uuid.NewString()
		}
		blob, err := compiler.NewPredictionSerializer().Serialize(&p)
		if err != nil {
			log.Printf("Failed to marshal prediction, with error: %v\n", err)
		}
		builder.addRow(p.UUID, blob, p.CreatedAt, p.PostedAt)
	}
	sql, args := builder.build()
	_, err := s.db.Exec(sql, args...)
	return err
}
