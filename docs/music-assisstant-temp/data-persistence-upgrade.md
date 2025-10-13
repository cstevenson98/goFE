# Data Persistence Upgrade Guide

## Overview

This guide documents the comprehensive upgrade of the Muse AI file management system from simple filesystem storage to a robust, containerized data persistence solution using PostgreSQL and MinIO.

## Architecture Changes

### Before (Current System)
- **Document Storage**: In-memory maps with filesystem backup
- **File Management**: Direct filesystem operations
- **Compilation Results**: Temporary files
- **No Versioning**: Files overwritten on update
- **No Metadata**: Limited document information

### After (New System)
- **Document Storage**: PostgreSQL database with full ACID compliance
- **File Management**: MinIO object storage with S3-compatible API
- **Compilation Results**: Tracked with detailed metadata
- **Version Control**: Full document and file versioning
- **Rich Metadata**: Tags, compilation history, user management

## New Features

### 🎯 **Enhanced File Management**
- **Persistent Storage**: All data survives container restarts
- **Version History**: Track changes to documents over time
- **File Organization**: Structured storage with metadata
- **Compilation Tracking**: Detailed compilation results and timing

### 🔍 **Advanced Search & Organization**
- **Tag-based Organization**: Categorize documents with tags
- **Full-text Search**: Search through document content
- **Metadata Filtering**: Filter by status, creation date, etc.
- **Project Management**: Organize documents into projects

### 👥 **Multi-User Support**
- **User Management**: Support for multiple users
- **Document Sharing**: Share documents with permissions
- **Activity Logging**: Track user actions and changes
- **Access Control**: Role-based permissions

### 📊 **Analytics & Reporting**
- **Compilation Statistics**: Track success rates and performance
- **Usage Analytics**: Monitor document creation and usage
- **File Size Tracking**: Monitor storage usage
- **Performance Metrics**: Compilation time analysis

## Database Schema

### Core Tables

#### Users
- User accounts with authentication
- Profile information and preferences
- Creation and modification timestamps

#### Projects
- Organize documents into projects
- User-owned with description and metadata
- Hierarchical organization support

#### Documents
- LilyPond documents with full versioning
- Status tracking (draft, compiling, compiled, error)
- Rich metadata including tags and JSON fields
- Compilation timing and results

#### Files
- All generated files (.ly, .pdf, .svg, .midi)
- Checksums for integrity verification
- MIME type detection and storage
- MinIO path references

#### Compilation Results
- Detailed compilation logs
- Error and warning tracking
- Performance metrics
- Version-specific results

## MinIO Object Storage

### File Organization
```
muse-ai-files/
├── documents/
│   ├── {document-uuid}/
│   │   ├── v1/
│   │   │   ├── ly/
│   │   │   │   └── document.ly
│   │   │   ├── pdf/
│   │   │   │   └── document.pdf
│   │   │   ├── svg/
│   │   │   │   └── document.svg
│   │   │   └── midi/
│   │   │       └── document.mid
│   │   ├── v2/
│   │   │   └── ...
│   │   └── ...
│   └── ...
```

### Features
- **Versioning**: Each document version stored separately
- **Checksums**: SHA-256 integrity verification
- **Metadata**: Rich file metadata storage
- **Presigned URLs**: Secure temporary file access
- **Bulk Operations**: Efficient file management

## API Enhancements

### New Endpoints

#### Document Management
- `GET /api/documents` - List documents with filtering
- `POST /api/documents` - Create new document
- `GET /api/documents/{id}` - Get document with files
- `PUT /api/documents/{id}` - Update document (new version)
- `DELETE /api/documents/{id}` - Delete document and files
- `POST /api/documents/{id}/compile` - Compile document
- `GET /api/documents/{id}/pdf` - Download PDF
- `GET /api/documents/{id}/files` - List all files for document

#### File Management
- `GET /api/files/{id}` - Get file metadata
- `GET /api/files/{id}/download` - Download file
- `GET /api/files/{id}/url` - Get presigned URL
- `DELETE /api/files/{id}` - Delete specific file

#### Analytics
- `GET /api/analytics/documents` - Document statistics
- `GET /api/analytics/compilation` - Compilation metrics
- `GET /api/analytics/storage` - Storage usage
- `GET /api/analytics/users` - User activity

### Enhanced Responses

#### Document Response
```json
{
  "id": "uuid",
  "project_id": "uuid",
  "user_id": "uuid",
  "title": "My Symphony",
  "content": "\\version \"2.22.1\"...",
  "status": "compiled",
  "version": 3,
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T12:00:00Z",
  "last_compiled_at": "2024-01-01T12:00:00Z",
  "compilation_duration": 1250,
  "tags": ["classical", "symphony", "major"],
  "files": [
    {
      "id": "uuid",
      "file_type": "ly",
      "file_name": "document.ly",
      "file_size": 1024,
      "mime_type": "text/plain",
      "checksum": "sha256...",
      "created_at": "2024-01-01T12:00:00Z"
    },
    {
      "id": "uuid",
      "file_type": "pdf",
      "file_name": "document.pdf",
      "file_size": 25600,
      "mime_type": "application/pdf",
      "checksum": "sha256...",
      "created_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```

## Deployment Guide

### Prerequisites
- Docker and Docker Compose
- At least 2GB RAM
- 10GB+ disk space for storage

### Environment Variables
```bash
# Database
DATABASE_URL=postgres://muse_ai:muse_ai_pass@postgres:5432/muse_ai?sslmode=disable

# MinIO
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=muse_ai
MINIO_SECRET_KEY=muse_ai_pass
MINIO_BUCKET=muse-ai-files

# Application
GO_ENV=production
```

### Quick Start
```bash
# Clone the repository
git clone https://github.com/your-org/muse-ai
cd muse-ai

# Start all services
docker-compose up -d

# Check health
docker-compose ps
```

### Services

#### PostgreSQL
- **Port**: 5432
- **Database**: muse_ai
- **User**: muse_ai
- **Password**: muse_ai_pass
- **Volume**: postgres_data

#### MinIO
- **API Port**: 9000
- **Console Port**: 9001
- **User**: muse_ai
- **Password**: muse_ai_pass
- **Volume**: minio_data

#### Application
- **Port**: 8081
- **Dependencies**: PostgreSQL, MinIO
- **Health Check**: `/health`

## Migration Guide

### From Current System

1. **Backup Existing Data**
   ```bash
   # Backup current LilyPond files
   tar -czf lilypond_backup.tar.gz /path/to/current/files
   ```

2. **Deploy New System**
   ```bash
   docker-compose up -d
   ```

3. **Migrate Data** (Optional migration script)
   ```bash
   # Run migration script
   go run scripts/migrate.go
   ```

### Data Migration Script
```go
// scripts/migrate.go
package main

import (
    "encoding/json"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
    
    "github.com/cstevenson98/muse-ai/internal/storage"
)

func main() {
    // Connect to new database
    db, err := storage.NewDatabase(os.Getenv("DATABASE_URL"))
    if err != nil {
        panic(err)
    }
    defer db.Close()
    
    // Connect to MinIO
    minioStorage, err := storage.NewMinIOStorage(...)
    if err != nil {
        panic(err)
    }
    
    // Migrate files from old system
    err = filepath.WalkDir("/path/to/old/files", func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            return err
        }
        
        if strings.HasSuffix(path, ".ly") {
            // Read file content
            content, err := os.ReadFile(path)
            if err != nil {
                return err
            }
            
            // Create document in new system
            doc := &storage.Document{
                Title:   filepath.Base(path),
                Content: string(content),
                Status:  "draft",
                Version: 1,
            }
            
            // Save to database and MinIO
            // ... migration logic
        }
        
        return nil
    })
    
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Migration completed successfully")
}
```

## Performance Considerations

### Database Optimization
- **Indexes**: Optimized for common query patterns
- **Connection Pooling**: Efficient connection management
- **Query Optimization**: Minimal N+1 queries

### MinIO Optimization
- **Chunked Upload**: Large file handling
- **Parallel Operations**: Concurrent file operations
- **Caching**: Metadata caching for performance

### Application Optimization
- **Background Processing**: Async compilation
- **Rate Limiting**: API rate limiting
- **Compression**: Response compression

## Monitoring & Maintenance

### Health Checks
```bash
# Check all services
curl http://localhost:8081/health

# Check database
docker-compose exec postgres pg_isready -U muse_ai

# Check MinIO
curl http://localhost:9000/minio/health/live
```

### Backup Strategy
```bash
# Database backup
docker-compose exec postgres pg_dump -U muse_ai muse_ai > backup.sql

# MinIO backup
mc mirror muse-ai-files/ backup/files/
```

### Log Monitoring
```bash
# Application logs
docker-compose logs -f muse-ai-backend

# Database logs
docker-compose logs -f muse-ai-postgres

# MinIO logs
docker-compose logs -f muse-ai-minio
```

## Frontend Integration

### API Client Updates
The frontend will need updates to use the new API endpoints:

```typescript
// New API client methods
interface DocumentAPI {
  createDocument(data: CreateDocumentRequest): Promise<Document>;
  updateDocument(id: string, data: UpdateDocumentRequest): Promise<Document>;
  deleteDocument(id: string): Promise<void>;
  compileDocument(id: string): Promise<CompilationResult>;
  downloadPDF(id: string): Promise<Blob>;
  getFiles(id: string): Promise<FileInfo[]>;
}
```

### UI Enhancements
- **File Browser**: Show all generated files
- **Version History**: Display document versions
- **Compilation Status**: Real-time compilation tracking
- **Tag Management**: Tag-based organization
- **Search Interface**: Advanced search capabilities

## Security Considerations

### Database Security
- **Encrypted Connections**: SSL/TLS for database connections
- **Access Control**: Role-based database access
- **Audit Logging**: Track all database operations

### MinIO Security
- **Access Keys**: Secure key management
- **Bucket Policies**: Fine-grained access control
- **Presigned URLs**: Temporary file access

### Application Security
- **Authentication**: JWT-based authentication
- **Authorization**: Role-based access control
- **Input Validation**: Comprehensive input sanitization
- **Rate Limiting**: API rate limiting

## Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check database status
docker-compose exec postgres pg_isready -U muse_ai

# Check logs
docker-compose logs muse-ai-postgres

# Reset database
docker-compose down -v
docker-compose up -d
```

#### MinIO Connection Issues
```bash
# Check MinIO status
curl http://localhost:9000/minio/health/live

# Check bucket
docker-compose exec minio mc ls muse-ai-files

# Reset MinIO
docker-compose down -v
docker-compose up -d
```

#### Application Issues
```bash
# Check application logs
docker-compose logs muse-ai-backend

# Check health endpoint
curl http://localhost:8081/health

# Restart application
docker-compose restart muse-ai-backend
```

## Future Enhancements

### Phase 2 Features
- **Real-time Collaboration**: Live document editing
- **Advanced Search**: Elasticsearch integration
- **File Formats**: Support for more output formats
- **Backup Automation**: Automated backup strategies

### Phase 3 Features
- **Cloud Storage**: S3/Azure integration
- **CDN Integration**: Fast file delivery
- **Advanced Analytics**: Machine learning insights
- **Mobile App**: React Native mobile client

## Support

For questions or issues:
- **Documentation**: [docs/](./docs/)
- **Issues**: [GitHub Issues](https://github.com/your-org/muse-ai/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-org/muse-ai/discussions)

## License

This project is licensed under the MIT License. See [LICENSE](../LICENSE) for details. 