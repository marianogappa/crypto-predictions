package statestorage

import (
	"fmt"
	"strings"
)

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

type filterable interface {
	filter() (string, []interface{})
}

type pgWhereBuilder struct {
	args      []interface{}
	filterArr []string
}

func (b *pgWhereBuilder) addFilter(f filterable) *pgWhereBuilder {
	s, args := f.filter()
	if s == "" {
		return b
	}
	b.filterArr = append(b.filterArr, s)
	b.args = append(b.args, args...)
	return b
}

func (b *pgWhereBuilder) addFilters(fs []filterable) *pgWhereBuilder {
	for _, f := range fs {
		b.addFilter(f)
	}
	return b
}

func (b *pgWhereBuilder) build() (string, []interface{}) {
	count := 1
	for i := range b.filterArr {
		for strings.Index(b.filterArr[i], "∆") != -1 {
			b.filterArr[i] = strings.Replace(b.filterArr[i], "∆", fmt.Sprintf("$%v", count), 1)
			count++
		}
	}

	if len(b.filterArr) == 0 {
		b.filterArr = append(b.filterArr, "$1 = $1")
		b.args = append(b.args, 1)
	}
	return strings.Join(b.filterArr, " AND "), b.args
}
