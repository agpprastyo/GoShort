# URL Shortener Service API

A RESTful API service for URL shortening built with Go and Fiber framework. This backend-only service provides endpoints for users to create and manage shortlinks, while giving admins moderation capabilities.

## Features

### User Capabilities
- **Create shortlinks** with customizable options:
    - Custom shortcodes
    - Expiration dates
    - Click limits
- **Manage your own shortlinks** via API
- **Full CRUD operations** for your own shortlinks

### Admin Capabilities
- **Moderation tools**:
    - Activate/deactivate any user's shortlink
    - Delete inappropriate or malicious shortlinks
- **Comprehensive audit logging** for all administrative actions

## Tech Stack

- **Backend**: Go with Fiber framework
- **Database**: PostgreSQL with SQLC for type-safe queries
- **Authentication**: JWT-based auth system
- **Logging**: Structured logging

## Architecture

The project follows clean architecture principles with clear separation of concerns:

```
├── api/            # API routes and handlers
├── config/         # Configuration management
├── db/
│   ├── migrations/ # Database migrations
│   └── sqlc/       # Generated database access code
├── internal/
│   ├── handler/    # Request handlers
│   ├── middleware/ # HTTP middleware components
│   ├── models/     # Domain models
│   ├── repository/ # Data access layer
│   └── service/    # Business logic layer
├── pkg/            # Reusable packages
└── cmd/            # Application entry points
```

## Key Implementation Details

- **Role-based access control** for users and admins
- **Comprehensive audit logging** for administrative actions
- **Customizable shortlinks** with validation
- **Secure JWT authentication**

## Deployment

The application is containerized with Docker for easy deployment and local development.

## Getting Started

```bash
# Clone the repository
git clone https://github.com/yourusername/urlshortener.git

# Navigate to project directory
cd urlshortener

# Start the development environment
docker-compose up -d

# Run migrations
make migrate-up

# Start the application
make run
```

## API Documentation

The API follows RESTful principles with the following endpoints:

### User Endpoints
- `POST /api/links` - Create new shortlink
- `GET /api/links` - List user's links
- `GET /api/links/:id` - Get link details
- `PUT /api/links/:id` - Update link
- `DELETE /api/links/:id` - Delete link

### Admin Endpoints
- `PATCH /api/admin/links/:id/status` - Toggle link active status
- `DELETE /api/admin/links/:id` - Delete any user's link

## Future Enhancements

- Web dashboard interface
- Analytics for link performance
- API key generation for programmatic access
- Custom domain support
- QR code generation for links

## License

MIT License