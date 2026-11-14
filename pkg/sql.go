package pkg

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

const MaxPaginationLimit = 500

var validColumnRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type NamedQueryer interface {
	NamedQueryContext(ctx context.Context, query string, arg interface{}) (*sqlx.Rows, error)
}

func extractIntPtr(v interface{}) *int {
	switch val := v.(type) {
	case *int:
		return val
	case int:
		return &val
	default:
		return nil
	}
}

func extractStringPtr(v interface{}) *string {
	switch val := v.(type) {
	case *string:
		return val
	case string:
		return &val
	default:
		return nil
	}
}

func SelectWithPagination(ctx context.Context, db NamedQueryer, query string, args map[string]interface{}) (*sqlx.Rows, error) {
	if !strings.HasSuffix(query, " ") {
		query += " "
	}

	queryArgs := make(map[string]interface{}, len(args)+2)
	for k, v := range args {
		queryArgs[k] = v
	}

	limit := extractIntPtr(queryArgs["limit"])
	page := extractIntPtr(queryArgs["page"])
	sort := extractStringPtr(queryArgs["sort"])

	if sort != nil && *sort != "" {
		sorts := strings.Split(*sort, ",")
		var orderClauses []string

		for _, part := range sorts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			parts := strings.Fields(part)
			if len(parts) == 0 {
				continue
			}

			field := parts[0]

			if !validColumnRegex.MatchString(field) {
				continue
			}

			direction := "ASC"

			if len(parts) > 1 && strings.ToUpper(parts[1]) == "DESC" {
				direction = "DESC"
			}

			orderClauses = append(orderClauses, fmt.Sprintf("%s %s", field, direction))
		}

		if len(orderClauses) > 0 {
			query += "ORDER BY " + strings.Join(orderClauses, ", ") + " "
		}
	}

	if limit != nil && *limit > 0 {
		safeLimit := *limit
		if safeLimit > MaxPaginationLimit {
			safeLimit = MaxPaginationLimit
		}

		safePage := 1
		if page != nil && *page > 0 {
			safePage = *page
		}

		offset := (safePage - 1) * safeLimit
		query += "LIMIT :limit OFFSET :offset"

		queryArgs["limit"] = safeLimit
		queryArgs["offset"] = offset
	} else {
		// Prevent unbounded full-table dumps when pagination params are omitted or invalid
		query += "LIMIT :limit OFFSET :offset"
		queryArgs["limit"] = MaxPaginationLimit
		queryArgs["offset"] = 0
	}

	rows, err := db.NamedQueryContext(ctx, query, queryArgs)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
