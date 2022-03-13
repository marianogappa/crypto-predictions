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

func MustNewPostgresDBStateStorage() PostgresDBStateStorage {
	p, err := NewPostgresDBStateStorage()
	if err != nil {
		log.Fatal(err)
	}
	return p
}

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

func buildWhere(filters types.APIFilters) (string, []interface{}) {
	args := []interface{}{}
	filterArr := []string{}

	if len(filters.AuthorHandles) > 0 {
		params := []string{}
		for _, authorHandle := range filters.AuthorHandles {
			args = append(args, authorHandle)
			params = append(params, fmt.Sprintf("$%v", len(args)))
		}
		filterArr = append(filterArr, fmt.Sprintf("blob->>'postAuthor' IN (%v)", strings.Join(params, ", ")))
	}

	if len(filters.PredictionStateValues) > 0 {
		params := []string{}
		for _, rawPredictionStateValue := range filters.PredictionStateValues {
			if _, err := types.PredictionStateValueFromString(rawPredictionStateValue); err != nil {
				continue
			}
			args = append(args, rawPredictionStateValue)
			params = append(params, fmt.Sprintf("$%v", len(args)))
		}
		if len(params) > 0 {
			filterArr = append(filterArr, fmt.Sprintf("blob->'state'->>'value' IN (%v)", strings.Join(params, ", ")))
		}
	}

	if len(filters.PredictionStateStatus) > 0 {
		params := []string{}
		for _, rawPredictionStateStatus := range filters.PredictionStateStatus {
			if _, err := types.ConditionStatusFromString(rawPredictionStateStatus); err != nil {
				continue
			}
			args = append(args, rawPredictionStateStatus)
			params = append(params, fmt.Sprintf("$%v", len(args)))
		}
		if len(params) > 0 {
			filterArr = append(filterArr, fmt.Sprintf("blob->'state'->>'status' IN (%v)", strings.Join(params, ", ")))
		}
	}

	if len(filters.UUIDs) > 0 {
		params := []string{}
		for _, uuid := range filters.UUIDs {
			args = append(args, uuid)
			params = append(params, fmt.Sprintf("$%v", len(args)))
		}
		filterArr = append(filterArr, fmt.Sprintf("uuid IN (%v)", strings.Join(params, ", ")))
	}

	if len(filterArr) == 0 {
		filterArr = append(filterArr, "$1 = $1")
		args = append(args, 1)
	}
	return strings.Join(filterArr, "AND"), args
}

func buildOrderBy(orderBys []string) string {
	if len(orderBys) == 0 {
		orderBys = []string{types.CREATED_AT_DESC.String()}
	}
	resultArr := []string{}
	for _, orderBy := range orderBys {
		switch orderBy {
		case types.CREATED_AT_DESC.String():
			resultArr = append(resultArr, "created_at DESC")
		case types.CREATED_AT_ASC.String():
			resultArr = append(resultArr, "created_at ASC")
		case types.POSTED_AT_DESC.String():
			resultArr = append(resultArr, "posted_at DESC")
		case types.POSTED_AT_ASC.String():
			resultArr = append(resultArr, "posted_at ASC")
		}
	}
	return strings.Join(resultArr, ", ")
}

func (s PostgresDBStateStorage) GetPredictions(filters types.APIFilters, orderBys []string) ([]types.Prediction, error) {
	where, args := buildWhere(filters)
	orderBy := buildOrderBy(orderBys)

	sql := fmt.Sprintf("SELECT uuid, blob FROM predictions WHERE %v ORDER BY %v", where, orderBy)

	// TODO filter
	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.Prediction{}
	for rows.Next() {
		var (
			clUUID, clBlob []byte
		)
		err := rows.Scan(&clUUID, &clBlob)
		if err != nil {
			log.Printf("error reading predictions fields from db, with error: %v\n", err)
		}
		var pred types.Prediction
		if pred, err = compiler.NewPredictionCompiler(nil, nil).Compile(clBlob); err != nil {
			log.Printf("read corrupted prediction from db, with error: %v\n", err)
			continue
		}
		pred.UUID = string(clUUID)
		result = append(result, pred)
	}

	return result, nil
}

func (s PostgresDBStateStorage) UpsertPredictions(ps []*types.Prediction) ([]*types.Prediction, error) {
	if len(ps) == 0 {
		return ps, nil
	}
	builder := newPGUpsertManyBuilder([]string{"uuid", "blob", "created_at", "posted_at"}, "predictions", "uuid")
	for i, p := range ps {
		if p.UUID == "" {
			ps[i].UUID = uuid.NewString()
		}
		blob, err := compiler.NewPredictionSerializer().Serialize(p)
		if err != nil {
			log.Printf("Failed to marshal prediction, with error: %v\n", err)
		}
		builder.addRow(p.UUID, blob, p.CreatedAt, p.PostedAt)
	}
	sql, args := builder.build()
	_, err := s.db.Exec(sql, args...)
	return ps, err
}

func (s PostgresDBStateStorage) LogPredictionStateValueChange(c types.PredictionStateValueChange) error {
	_, err := s.db.Exec(`
		INSERT INTO prediction_state_value_change
		(prediction_uuid, state_value, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (%v) DO UPDATE SET created_at = EXCLUDED.created_at
		`, c.PredictionUUID, c.StateValue, c.CreatedAt)

	return err
}
