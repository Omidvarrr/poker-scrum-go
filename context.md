# Planning Poker Application - Complete Context Documentation

## Overview
This is a **Planning Poker** web application built with Go, designed for agile teams to estimate story points collaboratively. Users can create rooms, join planning sessions, cast votes using Fibonacci-based scales, and see real-time results. The application features a modern dark UI, real-time WebSocket communication, and mobile-responsive design.

## Architecture

### Tech Stack
- **Backend**: Go 1.21 with Fiber v2.52.9 web framework
- **Database**: SQLite with GORM ORM for data persistence
- **Templates**: Templ v0.3.943 for type-safe HTML templating
- **WebSocket**: gorilla/websocket for real-time bidirectional communication
- **Authentication**: Google OAuth 2.0 with JWT tokens (HTTP-only cookies)
- **Frontend**: HTMX v1.9.10 for dynamic updates, vanilla JavaScript
- **Styling**: Custom CSS with mobile-first dark theme, responsive design
- **Deployment**: Docker with multi-stage builds, GitLab CI/CD, Darkube platform

### Project Structure
```
awesomeProject1/
├── internal/
│   ├── config/          # Configuration management
│   ├── database/        # Database connection and migrations
│   ├── dto/            # Data Transfer Objects
│   ├── handlers/       # HTTP request handlers
│   ├── models/         # Database models
│   ├── repositories/   # Data access layer
│   ├── services/       # Business logic layer
│   └── utils/          # Utility functions
├── web/
│   ├── static/         # CSS, images, fonts
│   └── templates/      # Templ template files
├── data/               # SQLite database and uploads
└── migrations/         # Database migration files
```

## Key Features

### 1. Authentication System
- **OAuth Integration**: Google OAuth for user authentication
- **JWT Tokens**: Secure session management
- **Profile Management**: Users can update names and avatars

### 2. Room Management
- **Room Creation**: Users can create planning rooms with names and avatars
- **Room Ownership**: Each room has an owner who can edit/delete the room
- **Room Settings**: Owners can modify room name, avatar, and delete rooms
- **No Membership System**: All authenticated users can join any room

### 3. Real-Time Voting
- **Planning Poker Scale**: 1, 2, 3, 5, 8, 13, 21, 34, ☕ (coffee break)
- **Live Voting**: WebSocket-based real-time vote casting
- **Vote Reveal**: Room owners/participants can reveal votes simultaneously
- **Average Calculation**: Automatically calculates average rounded to nearest Fibonacci value
- **Vote Reset**: Start new rounds by resetting votes

### 4. User Presence & Status
- **Live Rooms**: Home page shows only rooms with active users
- **Online Indicators**: Real-time display of who's online in rooms
- **Member Lists**: Live updating member lists in voting sessions
- **Connection Management**: WebSocket connection tracking and recovery

## Database Models (`internal/models/`)

### User (`user.go`)
```go
type User struct {
    ID           int    `gorm:"primaryKey" json:"id"`
    Email        string `gorm:"unique;not null" json:"email"`
    FirstName    string `json:"first_name"`
    LastName     string `json:"last_name"`
    Avatar       string `json:"avatar"`
    GoogleID     string `gorm:"unique" json:"google_id"`
    AccessToken  string `json:"-"` // Hidden from JSON
    RefreshToken string `json:"-"` // Hidden from JSON
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
```
**Relationships**: One-to-many with Votes, Rooms (as owner)

### Room (`room.go`)
```go
type Room struct {
    ID        string    `gorm:"primaryKey" json:"id"` // UUID
    Name      string    `gorm:"not null" json:"name"`
    Avatar    string    `json:"avatar"`
    OwnerId   string    `gorm:"not null" json:"owner_id"` // User ID as string
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
```
**Relationships**: Belongs-to User (owner), Has-many VoteSessions

### VoteSession (`vote.go`)
```go
type VoteSession struct {
    ID         string    `gorm:"primaryKey" json:"id"` // UUID
    RoomID     string    `gorm:"not null" json:"room_id"`
    IsRevealed bool      `gorm:"default:false" json:"is_revealed"`
    CreatedAt  time.Time `json:"created_at"`
    Votes      []Vote    `gorm:"foreignKey:SessionID" json:"votes,omitempty"`
}
```
**Relationships**: Belongs-to Room, Has-many Votes

### Vote (`vote.go`)
```go
type Vote struct {
    ID        int       `gorm:"primaryKey" json:"id"`
    RoomID    string    `gorm:"not null" json:"room_id"`
    UserID    int       `gorm:"not null" json:"user_id"`
    SessionID string    `gorm:"not null" json:"session_id"`
    VoteValue string    `json:"vote_value"` // "1", "2", "3", "5", "8", "13", "21", "34", "☕"
    VotedAt   time.Time `json:"voted_at"`
}
```
**Relationships**: Belongs-to User, VoteSession, Room

### Database Configuration
- **SQLite Database**: `data/app.db` for persistence
- **GORM Auto-Migration**: Automatic schema updates on startup
- **Connection Pooling**: Managed by GORM
- **Indexes**: Optimized queries for room/user lookups
- **Foreign Key Constraints**: Enforced referential integrity

## API Endpoints

### Authentication
- `GET /api/auth/google/login` - Initiate Google OAuth
- `GET /api/auth/google/callback` - OAuth callback
- `POST /api/auth/logout` - User logout

### Rooms (Protected)
- `GET /api/rooms` - List user's rooms
- `GET /api/rooms/live` - Get rooms with active users (HTML)
- `GET /api/rooms/all` - Get all rooms with status
- `POST /api/rooms` - Create new room
- `PUT /api/rooms/:id` - Update room (owners only)
- `DELETE /api/rooms/:id` - Delete room (owners only)
- `GET /api/rooms/:id/members` - Get room members
- `GET /api/rooms/:id/settings` - Get room settings drawer (owners only)

### Profile (Protected)
- `GET /api/profile` - Get user profile
- `PUT /api/profile` - Update profile
- `POST /api/profile/complete` - Complete profile setup

## WebSocket Real-Time Communication (`internal/handlers/websocket_handler.go`)

### WebSocketHandler Architecture
**Purpose**: Manages all real-time WebSocket connections and message routing

**Connection Management**:
- `clients map[string]map[*websocket.Conn]int` - Room-based connection tracking
- `globalClients map[*websocket.Conn]bool` - Global update subscribers (home page)
- Thread-safe operations with `sync.RWMutex`
- Automatic cleanup on disconnection

### WebSocket Message Events

#### Client → Server Events
- **`join_room`** - Join a voting room
  ```json
  {
    "type": "join_room",
    "room_id": "room-uuid",
    "data": {"user_id": 123}
  }
  ```
  - Adds user to room's active connections
  - Updates ActiveConnections global state
  - Broadcasts user presence to room members

- **`join_global_updates`** - Subscribe to global room updates (home page)
  - Used by home page to receive live room status updates
  - No authentication required for viewing live rooms

- **`cast_vote`** - Submit a vote
  ```json
  {
    "type": "cast_vote",
    "room_id": "room-uuid", 
    "data": {"vote_value": "5"}
  }
  ```
  - Validates user authentication
  - Creates/updates vote in current session
  - Supports vote removal with empty value

- **`reveal_votes`** - Reveal all votes (authorized users)
  - Triggers vote revelation for current session
  - Calculates and broadcasts results

- **`reset_votes`** - Start new voting round
  - Creates new voting session
  - Clears all previous votes

#### Server → Client Events
- **`vote_cast`** - Someone cast a vote
  ```json
  {
    "type": "vote_cast",
    "room_id": "room-uuid",
    "data": {"user_id": 123, "vote": "5"}
  }
  ```

- **`votes_revealed`** - Votes were revealed
  - Triggers UI refresh to show results and averages

- **`votes_reset`** - Voting round reset
  - Clears UI state for new voting round

- **`user_joined_room`** - User joined a room
  - Updates member lists and online indicators

- **`user_left_room`** - User left a room
  - Updates presence indicators
  - **Critical**: Always broadcast even when last user leaves (fixed bug)

- **`online_users_update`** - User presence updates
  ```json
  {
    "type": "online_users_update", 
    "online_users": [123, 456, 789]
  }
  ```

### Connection Lifecycle Management
- **Connection Establishment**: WebSocket upgrade with authentication validation
- **Auto-Reconnection**: Client-side reconnection logic with exponential backoff
- **Proper Cleanup**: `shouldReconnect` flag prevents reconnection on navigation
- **Memory Management**: Automatic removal of dead connections and empty rooms

## Detailed Service Layer Architecture

### RoomService (`internal/services/room_service.go`)
**Purpose**: Manages all room-related business logic and operations

**Key Methods**:
- `CreateRoom(userID int, request dto.CreateRoomRequest) (dto.RoomResponse, error)`
  - Creates new planning rooms with unique UUID
  - Handles optional image uploads to `data/uploads/rooms/`
  - Sets creator as room owner
- `GetRoomsList(userID int) (dto.RoomListResponse, error)`
  - Returns categorized rooms: user's owned rooms vs. other rooms
  - Implements ownership-based filtering
- `UpdateRoom(userID int, roomID string, request dto.UpdateRoomRequest) error`
  - Allows owners to modify room name and avatar
  - Enforces ownership validation before updates
- `DeleteRoom(userID int, roomID string) error`
  - Permanent room deletion (owners only)
  - Validates ownership before deletion
- `GetRoomMembers(roomID string, currentUserID int, baseURL string) (dto.RoomMembersResponse, error)`
  - Returns current room members with online status
  - Prioritizes current user in member list
  - Integrates with ActiveConnections for real-time presence
- `GetAllRoomsWithStatus(userID int, baseURL string) ([]dto.RoomResponse, error)`
  - Returns only "live" rooms (rooms with active users)
  - Uses ActiveConnections to filter active rooms
  - Powers the home page live rooms display

### VoteService (`internal/services/vote_service.go`)
**Purpose**: Handles all voting logic, session management, and result calculations

**Key Methods**:
- `CastVote(userID int, roomID string, request dto.VoteRequest) error`
  - Allows users to cast/update votes in active sessions
  - Supports vote removal (empty vote value)
  - Creates new session if none exists
- `RevealVotes(userID int, roomID string) error`
  - Reveals all votes in current session
  - Calculates averages and rounds to nearest Fibonacci value
  - Broadcasts results via WebSocket
- `ResetVotes(userID int, roomID string) error`
  - Creates new voting session
  - Clears all previous votes
  - Notifies all participants via WebSocket
- `GetCurrentVoteSession(roomID string, userID int) (dto.VoteResponse, error)`
  - Returns current session state with votes
  - Shows/hides votes based on reveal status
  - Includes user's own vote and voting statistics

**Fibonacci Calculation Logic**:
- Supports values: 1, 2, 3, 5, 8, 13, 21, 34, ☕ (coffee break)
- Non-numeric votes (coffee) excluded from averages
- Automatic rounding to nearest Fibonacci value
- Handles edge cases (no votes, all coffee votes)

### AuthService (`internal/services/auth_service.go`)
**Purpose**: Manages authentication and authorization

**Key Methods**:
- `ValidateGoogleToken(tokenString string) (*dto.GoogleUserInfo, error)`
  - Validates Google OAuth tokens
  - Extracts user information from Google API
- `CreateOrUpdateUser(googleInfo *dto.GoogleUserInfo) (*models.User, error)`
  - Creates new users or updates existing ones
  - Handles user profile synchronization
- JWT token generation and validation
- Session management with HTTP-only cookies

### ProfileService (`internal/services/profile_service.go`)
**Purpose**: User profile management and completion

**Key Methods**:
- `GetProfile(userID int) (dto.ProfileResponse, error)`
  - Returns user profile with avatar URL conversion
- `UpdateProfile(userID int, request dto.UpdateProfileRequest) error`
  - Updates user first/last name and avatar
  - Handles file uploads for avatar images
- `CompleteProfile(userID int, request dto.CompleteProfileRequest) error`
  - Completes initial profile setup after OAuth
  - Sets user names and optional avatar

### ActiveConnections (Global Singleton)
**Purpose**: Real-time WebSocket connection and presence management

**Type**: `RoomConnections` struct with thread-safe operations

**Key Methods**:
- `AddUser(roomID string, userID int)`
  - Tracks user joining a room
  - Thread-safe with mutex locking
- `RemoveUser(roomID string, userID int)`
  - Removes user from room tracking
  - Auto-cleans empty rooms
- `GetActiveRoomIDs() []string`
  - Returns list of rooms with active users
  - Powers the "live rooms" functionality
- `GetUsersInRoom(roomID string) []int`
  - Returns all active user IDs in a specific room
  - Used for member lists and presence indicators

**Data Structure**:
```go
type RoomConnections struct {
    mu    sync.RWMutex
    rooms map[string]map[int]struct{} // roomID -> set of userIDs
}
```

### GreetingService (`internal/services/greeting_service.go`)
**Purpose**: Provides personalized welcome messages and quotes

**Key Methods**:
- `GetGreeting(userID int) (dto.GreetingResponse, error)`
  - Returns personalized greeting with user's name
  - Includes inspirational quotes for planning sessions
  - Handles both authenticated and anonymous users

## UI Components

### Templates (Templ)
- **Pages**: Home, Login, Rooms, Profile, Vote
- **Components**: RoomCard, RoomSettings, User profiles
- **Base**: Common layout and navigation

### Key UI Features
- **Responsive Design**: Mobile-first with 4-column room grid on desktop
- **Dark Theme**: Custom dark UI with blue accents
- **Real-time Updates**: HTMX for seamless updates
- **Interactive Elements**: Vote buttons, settings drawers, confirmation dialogs

## Development Workflow & Deployment

### Building & Running
```bash
# Development
templ generate    # Generate template files
go run main.go    # Start development server (port 8080)

# Production Build
go build -o bin/app .
./bin/app
```

### Configuration Management (`internal/config/`)
**Environment Variables**:
- `PORT` - Server port (default: 8080)
- `DATABASE_PATH` - SQLite database file path
- `JWT_SECRET` - JWT token signing secret
- `OAUTH_STATE_SECRET` - OAuth state validation secret
- `GOOGLE_CLIENT_ID` - Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth client secret

**Config Structure**:
```go
type Config struct {
    Port             string
    DatabasePath     string 
    JWTSecret        string
    OAuthStateSecret string
    GoogleClientID   string
    GoogleClientSecret string
}
```

### Docker Deployment
**Multi-stage Dockerfile**:
- **Builder stage**: Go 1.21-alpine with templ generation
- **Runtime stage**: Alpine Linux with ca-certificates and SQLite
- **Security**: Non-root user execution
- **Health checks**: Built-in endpoint monitoring
- **Optimizations**: Layer caching, minimal image size

**Build Process**:
```bash
docker build -t planning-poker .
docker run -p 8080:8080 planning-poker
```

### GitLab CI/CD Pipeline (`.gitlab-ci.yml`)
**Stages**:
1. **Build**: Docker image creation with BUILDKIT
2. **Deploy**: Automated deployment to Darkube platform

**Environment Variables**:
- `CI_REGISTRY_USER` / `CI_REGISTRY_PASSWORD` - Docker registry authentication  
- `DARKUBE_DEV_TOKEN` / `DARKUBE_DEV_APP_ID` - Darkube deployment credentials

**Workflow**:
- Push to `develop` or `main` → Build image → Deploy automatically
- Image tagging with commit SHA for traceability

### Database Migrations
- **Auto-Migration**: GORM handles schema updates on startup
- **Migration Files**: SQL migrations in `migrations/` directory  
- **Database File**: SQLite stored in `data/app.db`
- **Backup Strategy**: File-based SQLite backups

### File Upload Management
- **Storage Location**: `data/uploads/` directory
- **Supported Types**: Images (avatars, room images)
- **Security**: File type validation, random filename generation
- **Serving**: Static file serving at `/uploads/*` route
- **Organization**: Subdirectories by type (`users/`, `rooms/`)
- **Permissions**: Proper directory permissions in Docker

## Security & Authentication (`internal/utils/jwt.go`)

### Authentication Flow
1. **Google OAuth 2.0**: Users authenticate via Google OAuth
2. **JWT Token Generation**: Server generates access and refresh tokens
3. **HTTP-Only Cookies**: Tokens stored in secure, HTTP-only cookies
4. **Token Validation**: Middleware validates JWT tokens from cookies or Authorization header

### JWT Implementation Details
```go
type JWTClaims struct {
    UserID int `json:"user_id"`
    jwt.RegisteredClaims
}
```

**Token Management**:
- **Access Token**: 24-hour expiration
- **Refresh Token**: 7-day expiration  
- **Cookie Security**: HTTPOnly, SameSite=Lax, Secure in production
- **Token Extraction**: Supports both cookie and Bearer token authentication

### Authorization & Permissions
- **Room Ownership**: Only room owners can edit/delete rooms
- **Profile Security**: Users can only update their own profiles  
- **WebSocket Authentication**: Connection validation with JWT tokens
- **Endpoint Protection**: JWT middleware on all sensitive routes

### File Upload Security
- **Type Validation**: Restricted to image files only
- **Filename Sanitization**: Random UUID-based naming prevents conflicts
- **Directory Isolation**: Uploads isolated in dedicated directories
- **Size Limits**: Reasonable file size restrictions

## Performance & Optimization

### Real-Time Communication
- **WebSocket Connection Pooling**: Efficient connection management
- **Selective Broadcasting**: Targeted message delivery to relevant users
- **Connection Cleanup**: Automatic removal of dead connections
- **Memory Management**: Efficient cleanup of empty rooms and connections

### Database Performance
- **GORM Optimizations**: Efficient ORM usage with proper relationships
- **Query Optimization**: Strategic use of indexes and joins
- **Connection Pooling**: Managed database connection lifecycle
- **Vote Aggregation**: Optimized Fibonacci calculation algorithms

### Frontend Performance
- **Static Asset Serving**: CSS/JS served from static files with proper caching
- **Minimal JavaScript**: Vanilla JS with strategic HTMX usage
- **Partial Updates**: HTMX for targeted DOM updates instead of full reloads
- **Responsive Design**: Mobile-first CSS with efficient responsive breakpoints
- **Component Reusability**: Templ components with efficient rendering

### Caching & Asset Management
- **Template Compilation**: Pre-compiled Templ templates
- **Static File Caching**: Browser caching headers for static assets
- **Connection State**: Efficient in-memory connection tracking

## Recent Bug Fixes & Improvements

### WebSocket Connection Issues (Fixed)
**Problem**: Rooms remained "live" after last user left
**Root Cause**: WebSocket reconnection logic persisting after navigation
**Solution**: 
- Added `shouldReconnect` flag to control reconnection behavior
- Implemented proper cleanup on page unload and HTMX navigation
- Fixed `user_left_room` broadcast for last user leaving scenario

### Form Submission Issues (Fixed)
**Problem**: Multiple form submissions possible during pending requests
**Solution**:
- Implemented multi-layer protection with `isSubmitting` flag
- Added proper button disable/enable state management
- Enhanced error handling and user feedback

### UI/UX Improvements (Fixed)
**Problem**: Room tiles not centered on `/rooms` page
**Solution**: Changed `.room-list` CSS from `justify-content: flex-start` to `justify-content: center`

### Dialog Functionality (Fixed)  
**Problem**: Confirmation dialogs (logout, room deletion) not working
**Solution**: 
- Fixed templ script syntax for dynamic onclick handlers
- Implemented proper function scope and event binding
- Added comprehensive error handling and user feedback

### Authentication Token Issues (Fixed)
**Problem**: Room deletion returning 401 errors
**Solution**: 
- Ensured `credentials: 'include'` in fetch requests for HTTP-only cookies
- Fixed JWT middleware cookie extraction logic
- Proper error handling for authentication failures

## Architecture Decisions & Trade-offs

### Technology Choices
- **Go + Fiber**: High performance, simple deployment, strong typing
- **SQLite**: Simple deployment, file-based persistence, good for single-server apps
- **Templ**: Type-safe templates, no runtime template parsing errors
- **HTMX**: Server-side rendering with dynamic updates, minimal JavaScript
- **WebSockets**: Real-time communication without polling overhead

### Design Patterns
- **Clean Architecture**: Separation of concerns (handlers, services, repositories)
- **Repository Pattern**: Abstract data access layer
- **Service Layer**: Business logic isolation
- **DTO Pattern**: Data transfer objects for API boundaries
- **Singleton Pattern**: Global connection management (ActiveConnections)

### Limitations & Considerations
- **Single Server Deployment**: SQLite limits horizontal scaling
- **File-based Storage**: Not suitable for distributed deployments
- **No Database Migrations**: Relies on GORM auto-migration
- **Basic User Roles**: Only owner/participant distinction
- **OAuth Dependency**: Requires Google OAuth for authentication

## Conclusion
This Planning Poker application demonstrates modern Go web development with real-time features, clean architecture, and responsive design. It successfully implements complex features like WebSocket communication, OAuth authentication, and file uploads while maintaining code simplicity and deployment ease. The application is production-ready with proper error handling, security measures, and performance optimizations.