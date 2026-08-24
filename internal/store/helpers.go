package store

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"task212-specline/internal/model"
)

// mapSQLError 把底层 SQL 错误映射为业务错误。
func mapSQLError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("database lookup failed: %v", model.ErrNotFound)
	}
	return err
}

// isUniqueViolation 判断是否为 UNIQUE 约束冲突。
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	// modernc.org/sqlite 约束冲突错误文本包含 UNIQUE constraint failed。
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
