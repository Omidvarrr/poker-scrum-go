# Planning Application - Context Documentation

## Overview
This is a **Planning Poker** web application built with Go, designed for agile teams to estimate story points collaboratively. Users can create rooms, join planning sessions, cast votes using Fibonacci-based scales, and see real-time results.

## Architecture

### Tech Stack
- **Backend**: Go with Fiber web framework
- **Database**: SQLite with GORM ORM
- **Templates**: Templ for HTML templating
- **WebSocket**: Real-time communication for voting and user presence
- **Authentication**: OAuth (Google) with JWT tokens
- **Frontend**: HTMX for dynamic updates, vanilla JavaScript
- **Styling**: Custom CSS with mobile-first dark theme

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

## Database Models

### User
```go
type User struct {
    ID        int    `gorm:"primaryKey"`
    Email     string `gorm:"unique;not null"`
    FirstName string
    LastName  string
    Avatar    string
    // OAuth fields...
}
```

### Room
```go
type Room struct {
    ID      string `gorm:"primaryKey"`
    Name    string `gorm:"not null"`
    Avatar  string
    OwnerId string `gorm:"not null"`
}
```

### Vote & VoteSession
```go
type VoteSession struct {
    ID         string `gorm:"primaryKey"`
    RoomID     string
    IsRevealed bool
    CreatedAt  time.Time
    Votes      []Vote `gorm:"foreignKey:SessionID"`
}

type Vote struct {
    ID        int    `gorm:"primaryKey"`
    RoomID    string
    UserID    int
    SessionID string
    VoteValue string
    VotedAt   time.Time
}
```

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

## WebSocket Events

### Client → Server
- `join_room` - Join a voting room
- `join_global_updates` - Subscribe to global room updates
- `cast_vote` - Submit a vote
- `reveal_votes` - Reveal all votes (authorized users)
- `reset_votes` - Start new voting round

### Server → Client
- `vote_cast` - Someone cast a vote
- `votes_revealed` - Votes were revealed
- `votes_reset` - Voting round reset
- `user_joined_room` - User joined a room
- `user_left_room` - User left a room
- `online_users_update` - User presence updates

## Key Services

### RoomService
- Room CRUD operations
- Member management
- Owner permission checks
- Live room status tracking

### VoteService
- Voting session management
- Vote casting and validation
- Results calculation with Fibonacci rounding
- Vote reveal/reset functionality

### AuthService & ProfileService
- OAuth integration
- JWT token management
- User profile operations

### ActiveConnections (Global)
- WebSocket connection tracking
- Real-time user presence
- Room activity monitoring

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

## Development Workflow

### Building & Running
```bash
templ generate    # Generate template files
go run main.go    # Start development server
```

### Database Migrations
- Located in `migrations/` directory
- Auto-applied on startup via `database.RunMigrations()`
- SQLite database stored in `data/app.db`

### File Uploads
- User avatars and room images
- Stored in `data/uploads/` directory
- Served at `/uploads/*` route
- Automatic format detection and random naming

## Security Considerations

### Authentication
- OAuth-only authentication (no passwords stored)
- JWT tokens for session management
- Middleware protection for sensitive endpoints

### Authorization
- Room owners can edit/delete their rooms
- Users can only update their own profiles
- WebSocket connections validated with room permissions

### File Handling
- File uploads restricted to images
- Random filename generation prevents conflicts
- Upload directory isolation

## Performance Features

### Real-time Communication
- WebSocket connection pooling
- Automatic reconnection on disconnect
- Efficient broadcast messaging for room updates

### Database Optimization
- Indexed queries for room/user lookups
- Connection pooling via GORM
- Efficient vote aggregation

### Frontend Performance
- CSS/JS served from static files
- Minimal JavaScript footprint
- HTMX for partial page updates instead of full reloads

## Known Limitations & Future Improvements

### Current Limitations
- SQLite database (single-server deployment)
- No user roles beyond room ownership
- Limited voting scale customization

### Architectural Decisions
- **No Membership System**: Removed in favor of simple ownership model
- **Real-time First**: All interactions use WebSocket for immediate feedback
- **Template-based**: Server-side rendering with minimal client-side JavaScript
- **Mobile Responsive**: 4-column desktop grid scales to single column on mobile

This application demonstrates modern Go web development patterns with real-time features, clean architecture, and responsive design suitable for team collaboration tools.