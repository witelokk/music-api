# Music Backend API 🎵

A modern Kotlin-based music streaming backend built with Ktor framework, providing authentication, music metadata, and API endpoints for the MusicApp Android client.

## Features

- **🔐 User authentication** with JWT tokens and Google Sign-In integration
- **📧 Email verification** system with Mailgun integration
- **🎵 Music metadata** management (songs, artists, releases, playlists)
- **❤️ Favorites & following** system for user personalization
- **🔍 Advanced search** across songs, artists, and playlists
- **📱 Personalized home screen** with dynamic content layout
- **🗄️ PostgreSQL database** with Exposed ORM
- **⚡ Redis caching** for session management and performance
- **📚 Auto-generated API documentation** with Swagger UI
- **🐳 Docker support** with complete development environment

## Android App

This backend powers the [MusicApp](https://github.com/witelokk/music-app) - a modern Android music streaming application built with Jetpack Compose and Material 3 design.

## Tech Stack

- **Framework**: Ktor 3.0+ with Netty engine
- **Language**: Kotlin 2.1.0+
- **Database**: PostgreSQL 17 with Exposed ORM
- **Caching**: Redis for sessions and performance
- **Authentication**: JWT tokens, Google OAuth2
- **Email**: Mailgun for verification emails
- **Documentation**: OpenAPI 3.0 with Swagger UI
- **Deployment**: Docker & Docker Compose
- **Reverse Proxy**: Nginx with JWT validation

## Requirements

- Kotlin 2.1.0+
- PostgreSQL 17+
- Redis 8+
- Docker & Docker Compose (for containerized deployment)
- Mailgun account for email verification
- Google Cloud project for OAuth2

## Environment Variables

Set the following environment variables:

```bash
# Database
DATABASE_URL=postgresql://localhost:5432/music_db
DATABASE_USER=postgres
DATABASE_PASSWORD=your_password

# Redis
REDIS_URL=localhost:6379

# JWT
JWT_SECRET=your_jwt_secret_key

# Mailgun
MAILGUN_API_KEY=your_mailgun_api_key
MAILGUN_DOMAIN=your_mailgun_domain
MAILGUN_FROM=noreply@yourdomain.com
MAILGUN_REGION=us  # or eu

# Google OAuth2
GOOGLE_AUTH_AUDIENCE=your_google_client_id
```

## Quick Start with Docker

1. **Clone the repository**
   ```bash
   git clone https://github.com/witelokk/music-backend.git
   cd music-backend
   ```

2. **Set environment variables**
   ```bash
   export DATABASE_URL="postgresql://music-api_postgres/"
   export DATABASE_USER="postgres"
   export DATABASE_PASSWORD="postgres"
   export REDIS_URL="music-api_redis:6379"
   export JWT_SECRET="your_jwt_secret"
   export MAILGUN_API_KEY="your_mailgun_api_key"
   export MAILGUN_DOMAIN="your_mailgun_domain"
   export MAILGUN_FROM="noreply@yourdomain.com"
   export MAILGUN_REGION="us"
   export GOOGLE_AUTH_AUDIENCE="your_google_client_id"
   ```

3. **Start the services**
   ```bash
   docker-compose up -d
   ```

4. **Access the API**
   - API: http://localhost:8080
   - Swagger UI: http://localhost:8080/swagger

## Local Development

1. **Set up PostgreSQL and Redis**
   ```bash
   # Using Docker for local development
   docker run -d --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 postgres:17
   docker run -d --name redis -p 6379:6379 redis
   ```

2. **Set environment variables**
   ```bash
   export DATABASE_URL="postgresql://localhost:5432/music_db"
   export DATABASE_USER="postgres"
   export DATABASE_PASSWORD="postgres"
   export REDIS_URL="localhost:6379"
   export JWT_SECRET="your_jwt_secret"
   export MAILGUN_API_KEY="your_mailgun_api_key"
   export MAILGUN_DOMAIN="your_mailgun_domain"
   export MAILGUN_FROM="noreply@yourdomain.com"
   export MAILGUN_REGION="us"
   export GOOGLE_AUTH_AUDIENCE="your_google_client_id"
   ```

3. **Run the application**
   ```bash
   ./gradlew run
   ```

4. **Run tests**
   ```bash
   ./gradlew test
   ```

## API Endpoints

### Authentication
- `POST /auth/login` - Email-based login
- `POST /auth/google` - Google OAuth2 login
- `POST /auth/refresh` - Refresh JWT token
- `POST /auth/logout` - Logout and invalidate token

### User Management
- `POST /users` - Create new user
- `GET /users/me` - Get current user profile
- `PUT /users/me` - Update user profile

### Verification
- `POST /verification/send` - Send verification email
- `POST /verification/verify` - Verify email code

### Music Content
- `GET /songs` - List songs with pagination
- `GET /songs/{id}` - Get song details
- `GET /artists` - List artists
- `GET /artists/{id}` - Get artist details
- `GET /releases` - List releases
- `GET /releases/{id}` - Get release details

### Playlists
- `GET /playlists` - List user playlists
- `POST /playlists` - Create new playlist
- `PUT /playlists/{id}` - Update playlist
- `DELETE /playlists/{id}` - Delete playlist
- `POST /playlists/{id}/songs` - Add song to playlist
- `DELETE /playlists/{id}/songs/{songId}` - Remove song from playlist

### Favorites & Following
- `POST /favorites/songs` - Add song to favorites
- `DELETE /favorites/songs/{songId}` - Remove song from favorites
- `POST /followings/artists` - Follow artist
- `DELETE /followings/artists/{artistId}` - Unfollow artist

### Search
- `GET /search` - Search across songs, artists, and playlists

### Home Screen
- `GET /home/layout` - Get personalized home screen layout

## Project Structure

```
music-backend/
├── src/main/kotlin/com/witelokk/music/
│   ├── routes/                 # API route handlers
│   │   ├── AuthRoutes.kt       # Authentication endpoints
│   │   ├── UsersRoutes.kt      # User management
│   │   ├── SongsRoutes.kt      # Song endpoints
│   │   ├── ArtistsRoutes.kt    # Artist endpoints
│   │   ├── PlaylistsRoutes.kt  # Playlist management
│   │   └── ...                 # Other route modules
│   ├── models/                 # Data models and DTOs
│   ├── tables/                 # Database table definitions
│   ├── Application.kt          # Main application setup
│   ├── Database.kt             # Database connection
│   ├── Auth.kt                 # Authentication configuration
│   └── EmailSender.kt          # Email service interface
├── src/main/resources/
│   ├── application.yaml        # Application configuration
│   └── static/                 # Static files
├── docker-compose.yml          # Docker services
├── Dockerfile                  # Application container
├── nginx.conf                  # Nginx configuration
└── build.gradle.kts           # Build configuration
```

## Database Schema

The application uses PostgreSQL with the following main tables:
- `users` - User accounts and profiles
- `songs` - Song metadata and file information
- `artists` - Artist information
- `releases` - Album/EP/Single releases
- `playlists` - User-created playlists
- `playlist_songs` - Playlist-song relationships
- `favorites` - User favorite songs
- `followers` - Artist following relationships

## License

MIT License - see [LICENSE](LICENSE.txt) file for details.
