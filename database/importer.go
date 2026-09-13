package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	autoIncrement    = regexp.MustCompile(`(?i)\s+AUTO_INCREMENT(?:=\d+)?`)
	columnComment    = regexp.MustCompile(`(?i)\s+COMMENT\s+(?:'(?:''|[^'])*'|"(?:""|[^"])*")`)
	columnCharset    = regexp.MustCompile(`(?i)\s+CHARACTER\s+SET\s+[A-Za-z0-9_]+(?:\s+COLLATE\s+[A-Za-z0-9_]+)?`)
	columnCollation  = regexp.MustCompile(`(?i)\s+COLLATE\s+[A-Za-z0-9_]+`)
	currentTimestamp = regexp.MustCompile(`(?i)CURRENT_TIMESTAMP\(\)`)
	onUpdate         = regexp.MustCompile(`(?i)\s+ON\s+UPDATE\s+(?:CURRENT_TIMESTAMP(?:\(\))?)`)
	unsigned         = regexp.MustCompile(`(?i)\s+UNSIGNED\b`)
	integerType      = regexp.MustCompile(`(?i)\b(?:TINY|SMALL|MEDIUM|BIG)?INT(?:\(\d+\))?(\s|,|\))`)
	autoColumn       = regexp.MustCompile("(?is)(?:`|\\\")([A-Za-z0-9_]+)(?:`|\\\")\\s+[^,\\n]*\\bAUTO_INCREMENT\\b")
)

func Import(ctx context.Context, databases *Databases, root string) error {
	files := map[Name]string{Auth: "auth/auth.sql", Realm: "realm/realm.sql", World: "world/world.sql", DBC: "dbc/dbc.sql"}
	for _, name := range names {
		if err := ImportFile(ctx, databases.DB(name), filepath.Join(root, files[name])); err != nil {
			return fmt.Errorf("import %s database: %w", name, err)
		}
	}
	_, err := databases.DB(Auth).ExecContext(ctx, `UPDATE realmlist SET realm_name = 'Moreno.AlphaCore' WHERE realm_id = 1`)
	return err
}

func ImportFile(ctx context.Context, db *sql.DB, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return importSQL(ctx, db, string(data), path)
}

func importSQL(ctx context.Context, db *sql.DB, data, source string) error {
	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return err
	}
	transaction, err := db.BeginTx(ctx, nil)
	if err != nil {
		db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
		return err
	}
	for _, raw := range splitStatements(data) {
		statement := normalizeSQL(raw)
		if statement == "" {
			continue
		}
		if _, err := transaction.ExecContext(ctx, statement); err != nil {
			transaction.Rollback()
			db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
			return fmt.Errorf("%s: %w; statement: %s", source, err, statementPreview(statement))
		}
	}
	if err := transaction.Commit(); err != nil {
		db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
		return err
	}
	_, err = db.ExecContext(ctx, "PRAGMA foreign_keys = ON")
	return err
}

func statementPreview(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if len(value) > 240 {
		return value[:240] + "..."
	}
	return value
}

func splitStatements(data string) []string {
	var result []string
	var statement strings.Builder
	var quote byte
	lineComment, blockComment := false, false
	for index := 0; index < len(data); index++ {
		value := data[index]
		if lineComment {
			if value == '\n' {
				lineComment = false
				statement.WriteByte(' ')
			}
			continue
		}
		if blockComment {
			if value == '*' && index+1 < len(data) && data[index+1] == '/' {
				blockComment = false
				index++
				statement.WriteByte(' ')
			}
			continue
		}
		if quote != 0 {
			statement.WriteByte(value)
			if value == '\\' && quote == '\'' && index+1 < len(data) {
				index++
				statement.WriteByte(data[index])
				continue
			}
			if value == quote {
				if index+1 < len(data) && data[index+1] == quote {
					index++
					statement.WriteByte(data[index])
				} else {
					quote = 0
				}
			}
			continue
		}
		switch value {
		case '\'', '"', '`':
			quote = value
			statement.WriteByte(value)
		case '#':
			lineComment = true
		case '-':
			if index+1 < len(data) && data[index+1] == '-' {
				lineComment = true
				index++
			} else {
				statement.WriteByte(value)
			}
		case '/':
			if index+1 < len(data) && data[index+1] == '*' {
				blockComment = true
				index++
			} else {
				statement.WriteByte(value)
			}
		case ';':
			result = append(result, statement.String())
			statement.Reset()
		default:
			statement.WriteByte(value)
		}
	}
	if statement.Len() > 0 {
		result = append(result, statement.String())
	}
	return result
}

func normalizeSQL(statement string) string {
	statement = normalizeEscapes(strings.TrimSpace(statement))
	if statement == "" {
		return ""
	}
	lower := strings.ToLower(statement)
	for _, prefix := range []string{"alter table", "lock tables", "unlock tables", "set ", "start transaction", "commit", "use ", "create database", "drop database", "grant ", "flush "} {
		if strings.HasPrefix(lower, prefix) {
			return ""
		}
	}
	statement = strings.ReplaceAll(statement, "INSERT IGNORE INTO", "INSERT OR IGNORE INTO")
	statement = strings.ReplaceAll(statement, "insert ignore into", "INSERT OR IGNORE INTO")
	if strings.HasPrefix(lower, "create table") {
		statement = normalizeCreateTable(statement)
	}
	return strings.TrimSpace(statement)
}

func normalizeCreateTable(statement string) string {
	lower := strings.ToLower(statement)
	autoColumnName := ""
	if match := autoColumn.FindStringSubmatch(statement); len(match) == 2 {
		autoColumnName = match[1]
	}
	if index := strings.Index(lower, ") engine="); index >= 0 {
		statement = statement[:index+1]
	}
	statement = autoIncrement.ReplaceAllString(statement, "")
	statement = unsigned.ReplaceAllString(statement, "")
	statement = currentTimestamp.ReplaceAllString(statement, "CURRENT_TIMESTAMP")
	statement = integerType.ReplaceAllString(statement, "INTEGER$1")
	statement = columnCharset.ReplaceAllString(statement, "")
	statement = columnCollation.ReplaceAllString(statement, "")
	statement = onUpdate.ReplaceAllString(statement, "")
	lines := strings.Split(statement, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.ToLower(strings.TrimSpace(line))
		if strings.HasPrefix(trimmed, "key ") || strings.HasPrefix(trimmed, "unique key ") || strings.HasPrefix(trimmed, "fulltext key ") {
			continue
		}
		if autoColumnName != "" && strings.Contains(trimmed, "primary key") && strings.Contains(trimmed, strings.ToLower(autoColumnName)) {
			continue
		}
		filtered = append(filtered, line)
	}
	statement = columnComment.ReplaceAllString(strings.Join(filtered, "\n"), "")
	if autoColumnName != "" {
		column := regexp.MustCompile("(?i)(`|\\\")" + regexp.QuoteMeta(autoColumnName) + "(`|\\\")\\s+INTEGER\\s+NOT\\s+NULL")
		statement = column.ReplaceAllString(statement, "${1}"+autoColumnName+"${2} INTEGER PRIMARY KEY AUTOINCREMENT")
	}
	statement = strings.ReplaceAll(statement, ",\n)", "\n)")
	statement = strings.ReplaceAll(statement, ",\r\n)", "\r\n)")
	if strings.Contains(strings.ToLower(statement), "create table `characters`") {
		statement = strings.Replace(statement, "`account`", "`account_id`", 1)
	}
	return statement
}

func normalizeEscapes(statement string) string {
	var result strings.Builder
	var quote byte
	for index := 0; index < len(statement); index++ {
		value := statement[index]
		if quote == '\'' {
			if value == '\\' && index+1 < len(statement) {
				index++
				switch statement[index] {
				case '\'':
					result.WriteString("''")
				case '\\':
					result.WriteByte('\\')
				case 'n':
					result.WriteByte('\n')
				case 'r':
					result.WriteByte('\r')
				case 't':
					result.WriteByte('\t')
				case 'b':
					result.WriteByte('\b')
				case 'Z':
					result.WriteByte(0x1a)
				default:
					result.WriteByte(statement[index])
				}
				continue
			}
			result.WriteByte(value)
			if value == '\'' {
				if index+1 < len(statement) && statement[index+1] == '\'' {
					index++
					result.WriteByte(statement[index])
				} else {
					quote = 0
				}
			}
			continue
		}
		result.WriteByte(value)
		if value == '\'' {
			quote = value
		}
	}
	return result.String()
}
