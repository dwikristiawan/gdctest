package squtil

import "github.com/Masterminds/squirrel"

// PG is the standard Squirrel StatementBuilder configured for PostgreSQL ($1, $2, ...)
var PG = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
