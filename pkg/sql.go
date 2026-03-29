package pkg

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

var validColumnRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func SelectWithPagination(ctx context.Context, db *sqlx.DB, query string, args map[string]interface{}) (*sqlx.Rows, error) {
	if !strings.HasSuffix(query, " ") {
		query += " "
	}

	if _, ok := args["limit"]; !ok {
		args["limit"] = 0
	}

	if _, ok := args["page"]; !ok {
		args["page"] = 0
	}

	if _, ok := args["sort"]; !ok {
		args["sort"] = ""
	}

	limit := args["limit"].(*int)
	page := args["page"].(*int)
	sort := args["sort"].(*string)

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

	if page != nil && *page != 0 && limit != nil && *limit != 0 {
		offset := (*page - 1) * *limit
		query += "LIMIT :limit OFFSET :offset"

		if args == nil {
			args = make(map[string]interface{})
		}
		args["limit"] = limit
		args["offset"] = offset
	}

	rows, err := db.NamedQueryContext(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return rows, nil
}
