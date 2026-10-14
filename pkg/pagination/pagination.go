package pagination

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strconv"
	"strings"

	"gdctest/pkg/echo"

	"github.com/Masterminds/squirrel"
	"golang.org/x/sync/errgroup"
)

type Request struct {
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Order string `json:"order"` // Format: "status:asc,created_at:desc"
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type DatabaseExecutor interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Sanitize ensures Request has valid Page and Limit values.
func (r *Request) Sanitize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 {
		r.Limit = 10
	} else if r.Limit > 100 {
		r.Limit = 100
	}
}

// FromEcho membaca query parameters pagination dari HTTP context
func FromEcho(c echo.Context) Request {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	req := Request{
		Page:  page,
		Limit: limit,
		Order: c.QueryParam("order"),
	}
	req.Sanitize()
	return req
}

// Fetch adalah executor async yang menjalankan count dan data list query secara paralel lewat errgroup
func Fetch[T any](
	ctx context.Context,
	db DatabaseExecutor,
	baseQuery squirrel.SelectBuilder,
	req Request,
	allowedSortFields map[string]bool,
	defaultOrders []string,
	scanner func(rows *sql.Rows) (T, error),
) ([]T, *Meta, error) {

	// 1. Normalisasi limit dan offset
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 10
	} else if req.Limit > 100 {
		req.Limit = 100
	}
	offset := uint64((req.Page - 1) * req.Limit)

	// 2. Count Query (Subquery count memastikan seluruh WHERE pada baseQuery 100% ikut dihitung)
	countQuery := squirrel.Select("COUNT(*)").
		FromSelect(baseQuery.RemoveLimit().RemoveOffset(), "count_subquery").
		PlaceholderFormat(squirrel.Dollar)

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build count sql: %w", err)
	}

	// 3. List Data Query (Menambahkan limit, offset, dan priority ordering)
	orderClauses := parseOrder(req.Order, allowedSortFields, defaultOrders)
	listQuery := baseQuery.
		Limit(uint64(req.Limit)).
		Offset(offset).
		OrderBy(orderClauses...)

	listSQL, listArgs, err := listQuery.ToSql()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to build list sql: %w", err)
	}

	var (
		totalRows int64
		items     []T
	)

	// 4. Eksekusi Paralel Menggunakan errgroup
	g, gCtx := errgroup.WithContext(ctx)

	// Worker A: Menghitung total_rows
	g.Go(func() error {
		if err := db.QueryRowContext(gCtx, countSQL, countArgs...).Scan(&totalRows); err != nil {
			return fmt.Errorf("failed to execute count query: %w", err)
		}
		return nil
	})

	// Worker B: Mengambil data dan melakukan scanning
	g.Go(func() error {
		rows, err := db.QueryContext(gCtx, listSQL, listArgs...)
		if err != nil {
			return fmt.Errorf("failed to execute list query: %w", err)
		}
		defer rows.Close()

		res := make([]T, 0)
		for rows.Next() {
			item, err := scanner(rows)
			if err != nil {
				return fmt.Errorf("failed to scan row: %w", err)
			}
			res = append(res, item)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		items = res
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	// 5. Kalkulasi Metadata Pagination
	totalPages := int(math.Ceil(float64(totalRows) / float64(req.Limit)))
	if totalPages < 1 {
		totalPages = 1
	}

	meta := &Meta{
		Page:       req.Page,
		Limit:      req.Limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1,
	}

	return items, meta, nil
}

func parseOrder(raw string, allowed map[string]bool, defaults []string) []string {
	var clauses []string
	if raw != "" {
		for _, part := range strings.Split(raw, ",") {
			items := strings.Split(strings.TrimSpace(part), ":")
			col := strings.ToLower(items[0])
			if !allowed[col] {
				continue
			}
			dir := "DESC"
			if len(items) > 1 && strings.ToLower(items[1]) == "asc" {
				dir = "ASC"
			}
			clauses = append(clauses, col+" "+dir)
		}
	}
	if len(clauses) == 0 {
		clauses = defaults
	}
	clauses = append(clauses, "id DESC")
	return clauses
}
