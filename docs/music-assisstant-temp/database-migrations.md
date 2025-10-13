# Database Migrations

This document explains how to write and manage database migrations in the Muse AI project.

## Overview

The Muse AI project uses a custom Go-based migration system that:
- Automatically initializes the database schema on startup
- Supports idempotent migrations (can be run multiple times safely)
- Includes retry logic for reliable database connections
- Parses complex SQL statements including functions and triggers

## Migration System Architecture

### Components

1. **`internal/dbinit` package**: Handles database initialization and migration execution
2. **`scripts/init-db.sql`**: Main migration file containing the complete schema
3. **Docker integration**: Migration files are copied into the container during build

### How It Works

1. On backend startup, the `dbinit` package connects to the database
2. It checks if the database is already initialized (by looking for the `users` table)
3. If not initialized, it reads and executes the SQL migration file
4. Each SQL statement is parsed and executed individually
5. The system handles complex statements like functions, triggers, and stored procedures

## Writing Migrations

### Current Approach

Currently, all migrations are in a single file: `scripts/init-db.sql`. This file contains:

```sql
-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create tables
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- ... other columns
);

-- Create indexes
CREATE INDEX idx_users_username ON users(username);

-- Create functions
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create triggers
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Migration Structure

The migration system expects the following structure in SQL files:

1. **Extensions**: Enable required PostgreSQL extensions
2. **Tables**: Create all tables with proper constraints
3. **Indexes**: Create indexes for performance
4. **Functions**: Create stored procedures and functions
5. **Triggers**: Create triggers for automated operations
6. **Initial Data**: Insert default/seed data

### Best Practices

#### 1. Use Idempotent Statements

Always use `IF NOT EXISTS` or `IF EXISTS` clauses to make migrations idempotent:

```sql
-- Good: Idempotent
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL
);

-- Bad: Not idempotent
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL
);
```

#### 2. Use Proper Constraints

Define constraints to ensure data integrity:

```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'compiling', 'compiled', 'error')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### 3. Create Indexes for Performance

Add indexes for frequently queried columns:

```sql
-- Single column index
CREATE INDEX IF NOT EXISTS idx_documents_user_id ON documents(user_id);

-- Composite index
CREATE INDEX IF NOT EXISTS idx_documents_user_status ON documents(user_id, status);

-- Partial index
CREATE INDEX IF NOT EXISTS idx_documents_active ON documents(user_id) WHERE status != 'deleted';

-- GIN index for JSON/Array columns
CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING GIN(tags);
```

#### 4. Use Proper Data Types

Choose appropriate PostgreSQL data types:

```sql
-- UUIDs for primary keys
id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

-- Timestamps with timezone
created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

-- JSON for flexible data
metadata JSONB DEFAULT '{}',

-- Arrays for lists
tags TEXT[],

-- Proper varchar lengths
username VARCHAR(50) NOT NULL,
email VARCHAR(255) NOT NULL
```

#### 5. Add Comments for Documentation

Document your schema:

```sql
-- Users table stores user account information
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add table comment
COMMENT ON TABLE users IS 'User accounts for the Muse AI application';
COMMENT ON COLUMN users.password_hash IS 'Bcrypt hashed password';
```

## Function and Trigger Examples

### Automatic Timestamp Updates

```sql
-- Function to update the updated_at column
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to automatically update updated_at
CREATE TRIGGER update_users_updated_at 
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

### Data Validation Functions

```sql
-- Function to validate email format
CREATE OR REPLACE FUNCTION validate_email(email TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$';
END;
$$ language 'plpgsql';

-- Use in a constraint
ALTER TABLE users ADD CONSTRAINT valid_email CHECK (validate_email(email));
```

## Adding New Migrations

### Option 1: Modify init-db.sql (Current Approach)

1. Edit `scripts/init-db.sql`
2. Add your new tables, indexes, or functions
3. Use `IF NOT EXISTS` clauses for idempotency
4. Test locally before deploying

### Option 2: Future Migration System (Recommended)

For a more robust system, consider implementing:

```
scripts/
├── migrations/
│   ├── 001_initial_schema.sql
│   ├── 002_add_projects_table.sql
│   ├── 003_add_file_storage.sql
│   └── 004_add_search_indexes.sql
└── init-db.sql (deprecated)
```

## Testing Migrations

### Local Testing

1. **Fresh Database**: Test on a clean database
```bash
# Drop and recreate database
psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS muse_ai_test;"
psql -h localhost -U postgres -c "CREATE DATABASE muse_ai_test;"

# Run your application to test migrations
go run main.go
```

2. **Existing Database**: Test idempotency
```bash
# Run migrations multiple times to ensure idempotency
go run main.go
go run main.go
go run main.go
```

### Docker Testing

Test the migration system in Docker:

```bash
# Build the Docker image
docker build -t muse-ai-backend .

# Run with database
docker-compose up -d postgres
docker run --rm --network muse-ai_default \
  -e DATABASE_URL="postgres://muse_ai:password@postgres:5432/muse_ai?sslmode=disable" \
  muse-ai-backend
```

## Configuration

### Environment Variables

The migration system uses these environment variables:

```bash
# Required: Database connection string
DATABASE_URL="postgres://user:password@host:port/database?sslmode=disable"

# Optional: Custom migration file path (defaults to scripts/init-db.sql)
MIGRATION_PATH="custom/path/to/migrations.sql"

# Optional: Retry configuration
DB_RETRY_ATTEMPTS=5
DB_RETRY_DELAY=5s
```

### Programmatic Configuration

```go
initializer := dbinit.New(dbinit.InitOptions{
    DatabaseURL:    "postgres://...",
    MigrationsPath: "scripts/init-db.sql",
    RetryAttempts:  5,
    RetryDelay:     5 * time.Second,
})

if err := initializer.Initialize(); err != nil {
    log.Fatalf("Failed to initialize database: %v", err)
}
defer initializer.Close()
```

## Troubleshooting

### Common Issues

1. **Migration fails with "relation already exists"**
   - Add `IF NOT EXISTS` clause
   - Check for duplicate statements

2. **Function creation fails**
   - Use `CREATE OR REPLACE FUNCTION`
   - Check for proper `$$` delimiters

3. **Migration hangs**
   - Check for deadlocks
   - Ensure proper transaction handling

### Debug Logging

The migration system provides detailed logging:

```
Initializing database...
Attempting to connect to database (attempt 1/5)...
Successfully connected to database
Database already initialized, skipping migration
```

### Manual Migration

If needed, you can run migrations manually:

```bash
# Connect to database
psql -h localhost -U muse_ai -d muse_ai

# Run specific parts of the migration
\i scripts/init-db.sql
```

## Security Considerations

1. **Use parameterized queries** in application code
2. **Limit database user permissions** to only required operations
3. **Validate input data** before database operations
4. **Use SSL connections** in production (`sslmode=require`)
5. **Audit sensitive operations** with triggers

## Performance Tips

1. **Create indexes before inserting large amounts of data**
2. **Use `ANALYZE` after major data changes**
3. **Consider `VACUUM` for maintenance**
4. **Monitor query performance** with `EXPLAIN ANALYZE`

## Future Enhancements

Consider implementing these features:

1. **Versioned migrations**: Track migration versions in a table
2. **Rollback support**: Ability to rollback migrations
3. **Migration validation**: Syntax checking before execution
4. **Parallel execution**: Run independent migrations in parallel
5. **Migration hooks**: Pre/post migration scripts 