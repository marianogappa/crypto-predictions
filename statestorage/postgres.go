package statestorage

import (
	"database/sql"
	"embed"
	"fmt"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"

	// Storage engine
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	pq "github.com/lib/pq"
	"github.com/marianogappa/predictions/compiler"
	"github.com/marianogappa/predictions/serializer"
	"github.com/marianogappa/predictions/types"
	"github.com/rs/zerolog/log"
)

// PostgresConf is the configuration for the Postgres instance.
type PostgresConf struct {
	User     string
	Host     string
	Pass     string
	Port     string
	Database string
}

// PostgresDBStateStorage is the Postgres implementation of StateStorage.
type PostgresDBStateStorage struct {
	db    *sql.DB
	debug bool
}

//go:embed migrations/*.sql
var fs embed.FS

// MustNewPostgresDBStateStorage constructs a PostgresDBStateStorage. May fatal.
func MustNewPostgresDBStateStorage(c PostgresConf) *PostgresDBStateStorage {
	p, err := NewPostgresDBStateStorage(c)
	if err != nil {
		connStr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", c.User, c.Pass, c.Host, c.Port, c.Database)
		log.Fatal().Err(err).Msgf("An addressable postgres database is required. Currently looking for it in: %v. Configure these parameters via the PREDICTIONS_POSTGRES_ env variables described in the README.", connStr)
	}
	return &p
}

// NewPostgresDBStateStorage constructs a PostgresDBStateStorage.
func NewPostgresDBStateStorage(c PostgresConf) (PostgresDBStateStorage, error) {
	connStr := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", c.User, c.Pass, c.Host, c.Port, c.Database)

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

	log.Info().Str("url", connStr).Msgf("Connected to Postgres DB")
	return PostgresDBStateStorage{db: db}, nil
}

// SetDebug sets the debug logging setting across the storage layer.
func (s *PostgresDBStateStorage) SetDebug(debug bool) {
	s.debug = debug
}

// DB returns the DB for raw queries
func (s *PostgresDBStateStorage) DB() *sql.DB {
	return s.db
}

func predictionsBuildOrderBy(orderBys []string) string {
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
		case types.UUID_ASC.String():
			resultArr = append(resultArr, "uuid ASC")
		}
	}
	return strings.Join(resultArr, ", ")
}

// GetPredictions SELECTs predictions from the database.
func (s PostgresDBStateStorage) GetPredictions(filters types.APIFilters, orderBys []string, limit, offset int) ([]types.Prediction, error) {
	where, args := (&pgWhereBuilder{}).addFilters([]filterable{
		pgPredictionsAuthorHandles{filters.AuthorHandles},
		pgPredictionsAuthorURLs{filters.AuthorURLs},
		pgPredictionsDeleted{filters.Deleted},
		pgPredictionsHidden{filters.Hidden},
		pgPredictionsPaused{filters.Paused},
		pgPredictionsPredictionStateStatuses{filters.PredictionStateStatus},
		pgPredictionsPredictionStateValues{filters.PredictionStateValues},
		pgPredictionsUUIDs{filters.UUIDs},
		pgPredictionsURLs{filters.URLs},
		pgPredictionsTags{filters.Tags},
		pgGreaterThanUUID{filters.GreaterThanUUID},
		pgIncludeUIUnsupported{filters.IncludeUIUnsupported},
	}).build()

	orderBy := predictionsBuildOrderBy(orderBys)
	limitStr := ""
	if limit > 0 {
		limitStr = fmt.Sprintf(" LIMIT %v OFFSET %v", limit, offset)
	}

	sql := fmt.Sprintf("SELECT uuid, blob FROM predictions WHERE %v ORDER BY %v%v", where, orderBy, limitStr)

	if s.debug {
		log.Info().Msgf("PostgresDBStateStorage.GetPredictions: for filters %+v and orderBy %+v: %v\n", filters, orderBys, sql)
	}

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
			log.Info().Msgf("error reading predictions fields from db, with error: %v\n", err)
		}
		var pred types.Prediction
		if pred, _, err = compiler.NewPredictionCompiler(nil, nil).Compile(clBlob); err != nil {
			log.Info().Msgf("read corrupted prediction from db, with error: %v\n", err)
			continue
		}
		pred.UUID = string(clUUID)
		result = append(result, pred)
	}

	return result, nil
}

func accountsBuildOrderBy(orderBys []string) string {
	if len(orderBys) == 0 {
		orderBys = []string{types.ACCOUNT_FOLLOWER_COUNT_DESC.String()}
	}
	resultArr := []string{}
	for _, orderBy := range orderBys {
		switch orderBy {
		case types.ACCOUNT_CREATED_AT_DESC.String():
			resultArr = append(resultArr, "created_at DESC")
		case types.ACCOUNT_CREATED_AT_ASC.String():
			resultArr = append(resultArr, "created_at ASC")
		case types.ACCOUNT_FOLLOWER_COUNT_DESC.String():
			resultArr = append(resultArr, "follower_count DESC")
		}
	}
	return strings.Join(resultArr, ", ")
}

// GetAccounts SELECTs accounts from the database.
func (s PostgresDBStateStorage) GetAccounts(filters types.APIAccountFilters, orderBys []string, limit, offset int) ([]types.Account, error) {
	where, args := (&pgWhereBuilder{}).addFilters([]filterable{
		pgAccountsHandles{filters.Handles},
		pgAccountsURLs{filters.URLs},
	}).build()

	orderBy := accountsBuildOrderBy(orderBys)
	limitStr := ""
	if limit > 0 {
		limitStr = fmt.Sprintf(" LIMIT %v OFFSET %v", limit, offset)
	}

	sql := fmt.Sprintf("SELECT url, account_type, handle, follower_count, thumbnails, name, description, created_at, is_verified FROM accounts WHERE %v ORDER BY %v%v", where, orderBy, limitStr)

	if s.debug {
		log.Info().Msgf("PostgresDBStateStorage.GetAccounts: for filters %+v and orderBy %+v: %v\n", filters, orderBys, sql)
	}

	rows, err := s.db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []types.Account{}
	for rows.Next() {
		var (
			a          types.Account
			dbURL      string
			thumbnails []string
			createdAt  pq.NullTime
		)

		err := rows.Scan(&dbURL, &a.AccountType, &a.Handle, &a.FollowerCount, pq.Array(&thumbnails), &a.Name, &a.Description, &createdAt, &a.IsVerified)
		if err != nil {
			log.Info().Msgf("error reading account from db, with error: %v\n", err)
		}

		u, err := url.Parse(dbURL)
		if err != nil {
			log.Info().Msgf("error reading url from account from db, with error: %v\n", err)
		}
		a.URL = u

		for _, dbURL := range thumbnails {
			u, err := url.Parse(dbURL)
			if err != nil {
				log.Info().Msgf("error reading url from thumbnails from account from db, with error: %v\n", err)
			}
			a.Thumbnails = append(a.Thumbnails, u)
		}

		if createdAt.Valid {
			a.CreatedAt = &createdAt.Time
		}

		result = append(result, a)
	}

	return result, nil
}

// UpsertPredictions UPSERTs predictions to the database.
func (s PostgresDBStateStorage) UpsertPredictions(ps []*types.Prediction) ([]*types.Prediction, error) {
	if len(ps) == 0 {
		return ps, nil
	}

	builder := newPGUpsertManyBuilder([]string{"uuid", "blob", "created_at", "posted_at", "tags", "post_url"}, "predictions", "uuid")
	for i := range ps {
		if ps[i].UUID == "" {
			ps[i].UUID = uuid.NewString()
		}
		blob, err := serializer.NewPredictionSerializer(nil).Serialize(ps[i])
		if err != nil {
			log.Info().Msgf("Failed to marshal prediction, with error: %v\n", err)
		}
		builder.addRow(ps[i].UUID, blob, ps[i].CreatedAt, ps[i].PostedAt, pq.Array(ps[i].CalculateTags()), ps[i].PostUrl)
	}
	sql, args := builder.build()
	_, err := s.db.Exec(sql, args...)
	return ps, err
}

// PausePrediction sets a prediction to paused on the database. Paused predictions are visible but don't evolve.
func (s PostgresDBStateStorage) PausePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET paused = true WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// UnpausePrediction sets a prediction to unpaused on the database. Paused predictions are visible but don't evolve.
func (s PostgresDBStateStorage) UnpausePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET paused = false WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// HidePrediction sets a prediction to hidden on the database. Hidden predictions are invisible but still evolve.
func (s PostgresDBStateStorage) HidePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET hidden = true WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// UnhidePrediction sets a prediction to visible on the database. Hidden predictions are invisible but still evolve.
func (s PostgresDBStateStorage) UnhidePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET hidden = false WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// DeletePrediction sets a prediction to deleted on the database. Deleted predictions are invisible and don't evolve.
func (s *PostgresDBStateStorage) DeletePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET deleted = true WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// UndeletePrediction restores a deleted prediction on the database. Deleted predictions are invisible and don't evolve.
func (s *PostgresDBStateStorage) UndeletePrediction(uuid string) error {
	res, err := s.db.Exec("UPDATE predictions SET deleted = false WHERE uuid::text = $1", uuid)
	if err != nil {
		return err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return nil
	}
	if count == 0 {
		return fmt.Errorf("uuid not found: %v", uuid)
	}
	return nil
}

// UpsertAccounts UPSERTs accounts to the database.
func (s PostgresDBStateStorage) UpsertAccounts(as []*types.Account) ([]*types.Account, error) {
	if len(as) == 0 {
		return as, nil
	}

	builder := newPGUpsertManyBuilder([]string{"url", "account_type", "handle", "follower_count", "thumbnails", "name", "description", "created_at", "is_verified"}, "accounts", "url")
	for _, a := range as {
		thumbnails := []string{}
		for _, thumb := range a.Thumbnails {
			thumbnails = append(thumbnails, thumb.String())
		}
		builder.addRow(a.URL.String(), a.AccountType, a.Handle, a.FollowerCount, pq.Array(thumbnails), a.Name, a.Description, a.CreatedAt, a.IsVerified)
	}
	sql, args := builder.build()
	_, err := s.db.Exec(sql, args...)
	return as, err
}

// LogPredictionStateValueChange logs the fact that a prediction changed PredictionStateValue to the database.
func (s PostgresDBStateStorage) LogPredictionStateValueChange(c types.PredictionStateValueChange) error {
	_, err := s.db.Exec(`
		INSERT INTO prediction_state_value_change
		(prediction_uuid, state_value, created_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (prediction_uuid, state_value) DO UPDATE SET created_at = EXCLUDED.created_at
		`, c.PredictionUUID, c.StateValue, c.CreatedAt)

	return err
}

// PredictionInteractionExists checks the database to see if a predictions creation or finalization Tweet post happened.
func (s PostgresDBStateStorage) PredictionInteractionExists(predictionUUID, postURL, actionType string) (bool, error) {
	var exists bool
	res, err := s.db.Query(`
	SELECT EXISTS(SELECT * FROM prediction_interactions WHERE prediction_uuid = $1 AND post_url = $2 AND action_type = $3);
		`, predictionUUID, postURL, actionType)
	if err != nil {
		return false, err
	}
	res.Next()
	if err := res.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// InsertPredictionInteraction logs the fact that a Tweet was sent when a prediction was created or finalized.
func (s PostgresDBStateStorage) InsertPredictionInteraction(predictionUUID, postURL, actionType, interactionPostURL string) error {
	_, err := s.db.Query(`
	INSERT INTO prediction_interactions (uuid, prediction_uuid, post_url, action_type, interaction_post_url) VALUES ($1, $2, $3, $4, $5);
		`, uuid.NewString(), predictionUUID, postURL, actionType, interactionPostURL)
	if err != nil {
		return err
	}
	return nil
}

type pgPredictionsDeleted struct{ deleted *bool }

func (d pgPredictionsDeleted) filter() (string, []interface{}) {
	if d.deleted != nil && *d.deleted {
		return "deleted is true", []interface{}{}
	}
	if d.deleted != nil && !*d.deleted {
		return "deleted is not true", []interface{}{}
	}
	return "", nil
}

type pgPredictionsHidden struct{ hidden *bool }

func (d pgPredictionsHidden) filter() (string, []interface{}) {
	if d.hidden != nil && *d.hidden {
		return "hidden is true", []interface{}{}
	}
	if d.hidden != nil && !*d.hidden {
		return "hidden is not true", []interface{}{}
	}
	return "", nil
}

type pgPredictionsPaused struct{ paused *bool }

func (d pgPredictionsPaused) filter() (string, []interface{}) {
	if d.paused != nil && *d.paused {
		return "paused is true", []interface{}{}
	}
	if d.paused != nil && !*d.paused {
		return "paused is not true", []interface{}{}
	}
	return "", nil
}

type pgPredictionsAuthorHandles struct{ authorHandles []string }

func (f pgPredictionsAuthorHandles) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, authorHandle := range f.authorHandles {
		args = append(args, authorHandle)
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->>'postAuthor' IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsAuthorURLs struct{ authorURLs []string }

func (f pgPredictionsAuthorURLs) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, authorURL := range f.authorURLs {
		args = append(args, authorURL)
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->>'postAuthorURL' IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsPredictionStateValues struct{ predictionStateValues []string }

func (f pgPredictionsPredictionStateValues) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, rawPredictionStateValue := range f.predictionStateValues {
		if _, err := types.PredictionStateValueFromString(rawPredictionStateValue); err != nil {
			continue
		}
		args = append(args, rawPredictionStateValue)
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->'state'->>'value' IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsPredictionStateStatuses struct{ predictionStateStatuses []string }

func (f pgPredictionsPredictionStateStatuses) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, rawPredictionStateStatus := range f.predictionStateStatuses {
		if _, err := types.ConditionStatusFromString(rawPredictionStateStatus); err != nil {
			continue
		}
		args = append(args, rawPredictionStateStatus)
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->'state'->>'status' IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsUUIDs struct{ uuids []string }

func (f pgPredictionsUUIDs) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, uuid := range f.uuids {
		args = append(args, uuid)
	}
	if len(args) > 0 {
		return fmt.Sprintf("uuid::text IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsTags struct{ tags []string }

func (f pgPredictionsTags) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, tag := range f.tags {
		args = append(args, tag)
	}
	if len(args) > 0 {
		return fmt.Sprintf("tags && ARRAY[%v]", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgPredictionsURLs struct{ urls []string }

func (f pgPredictionsURLs) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, url := range f.urls {
		args = append(args, url)
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->>'postUrl' IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgAccountsHandles struct{ handles []string }

func (f pgAccountsHandles) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, handle := range f.handles {
		args = append(args, handle)
	}
	if len(args) > 0 {
		return fmt.Sprintf("handle IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgAccountsURLs struct{ urls []string }

func (f pgAccountsURLs) filter() (string, []interface{}) {
	args := []interface{}{}
	for _, url := range f.urls {
		args = append(args, url)
	}
	if len(args) > 0 {
		return fmt.Sprintf("url IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}

type pgGreaterThanUUID struct{ uuid string }

func (f pgGreaterThanUUID) filter() (string, []interface{}) {
	if f.uuid == "" {
		return "", nil
	}
	return "uuid > ∆", []interface{}{f.uuid}
}

type pgIncludeUIUnsupported struct{ includeUIUnsupported bool }

func (f pgIncludeUIUnsupported) filter() (string, []interface{}) {
	if f.includeUIUnsupported {
		return "", nil
	}
	args := []interface{}{}
	for predictionType := range types.UIUnsupportedPredictionTypes {
		args = append(args, predictionType.String())
	}
	if len(args) > 0 {
		return fmt.Sprintf("blob->>'type' NOT IN (%v)", strings.Join(strings.Split(strings.Repeat("∆", len(args)), ""), ", ")), args
	}
	return "", nil
}
