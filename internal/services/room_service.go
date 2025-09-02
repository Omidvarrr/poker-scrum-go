package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/models"
	"awesomeProject1/internal/repositories"
	"awesomeProject1/internal/utils"
	"errors"
	"strconv"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo repositories.RoomRepository
	userRepo repositories.UserRepository
}

func NewRoomService(
	roomRepo repositories.RoomRepository,
	userRepo repositories.UserRepository,
) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		userRepo: userRepo,
	}
}

func (rs *RoomService) CreateRoom(userID int, request dto.CreateRoomRequest) (dto.RoomResponse, error) {
	roomID := uuid.New().String()

	room := models.Room{
		ID:      roomID,
		Name:    request.Name,
		Avatar:  request.Avatar,
		OwnerId: strconv.Itoa(userID),
	}

	createdRoom, err := rs.roomRepo.CreateRoom(room)
	if err != nil {
		return dto.RoomResponse{}, err
	}

	if err != nil {
		return dto.RoomResponse{}, err
	}

	return dto.RoomResponse{
		ID:      createdRoom.ID,
		Name:    createdRoom.Name,
		Avatar:  createdRoom.Avatar,
		OwnerID: createdRoom.OwnerId,
		IsOwner: true,
	}, nil
}

func (rs *RoomService) GetRoomsList(userID int) (dto.RoomListResponse, error) {
	userIDStr := strconv.Itoa(userID)
	myRooms, err := rs.roomRepo.GetRoomsByOwner(userID)
	if err != nil {
		return dto.RoomListResponse{}, err
	}

	allRooms, err := rs.roomRepo.GetAllRooms()
	if err != nil {
		return dto.RoomListResponse{}, err
	}

	var myRoomResponses []dto.RoomResponse
	var otherRoomResponses []dto.RoomResponse

	for _, room := range myRooms {
		myRoomResponses = append(
			myRoomResponses, dto.RoomResponse{
				ID:      room.ID,
				Name:    room.Name,
				Avatar:  room.Avatar,
				OwnerID: room.OwnerId,
				IsOwner: true,
			},
		)
	}

	for _, room := range allRooms {
		if room.OwnerId != userIDStr {
			otherRoomResponses = append(
				otherRoomResponses, dto.RoomResponse{
					ID:      room.ID,
					Name:    room.Name,
					Avatar:  room.Avatar,
					OwnerID: room.OwnerId,
					IsOwner: room.OwnerId == userIDStr,
				},
			)
		}
	}

	return dto.RoomListResponse{
		MyRooms:    myRoomResponses,
		OtherRooms: otherRoomResponses,
	}, nil
}

func (rs *RoomService) UpdateRoom(userID int, roomID string, request dto.UpdateRoomRequest) error {
	isOwner, err := rs.roomRepo.IsUserOwner(roomID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only room owner can update room settings")
	}

	return rs.roomRepo.UpdateRoom(roomID, request.Name, request.Avatar)
}

func (rs *RoomService) DeleteRoom(userID int, roomID string) error {
	isOwner, err := rs.roomRepo.IsUserOwner(roomID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only room owner can delete room")
	}

	return rs.roomRepo.DeleteRoom(roomID)
}

func (rs *RoomService) GetRoomMembers(roomID string, currentUserID int, baseURL string) (dto.RoomMembersResponse, error) {
	activeUserIDs := ActiveConnections.GetUsersInRoom(roomID)

	// Create a set to track users we've added to avoid duplicates
	userMap := make(map[int]bool)
	var memberInfos []dto.RoomMemberInfo

	// Always include the current user first
	if currentUserID > 0 {
		user, err := rs.userRepo.GetUserById(currentUserID)
		if err == nil {
			isOnline := contains(activeUserIDs, currentUserID)
			memberInfos = append(memberInfos, dto.RoomMemberInfo{
				UserID:    user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				Avatar:    utils.ConvertToURL(user.Avatar, baseURL),
				IsOnline:  isOnline,
			})
			userMap[currentUserID] = true
		}
	}

	// Add other active users
	for _, userID := range activeUserIDs {
		if userMap[userID] {
			continue // Skip if already added
		}
		user, err := rs.userRepo.GetUserById(userID)
		if err != nil {
			continue
		}
		memberInfos = append(memberInfos, dto.RoomMemberInfo{
			UserID:    user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Avatar:    utils.ConvertToURL(user.Avatar, baseURL),
			IsOnline:  true,
		})
		userMap[userID] = true
	}

	return dto.RoomMembersResponse{
		Members:       memberInfos,
		OnlineMembers: activeUserIDs,
	}, nil
}

// Helper function to check if slice contains value
func contains(slice []int, value int) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func (rs *RoomService) GetAllRooms(userID int, baseURL string) ([]dto.RoomResponse, error) {
	allRooms, err := rs.roomRepo.GetAllRooms()
	if err != nil {
		return nil, err
	}

	var roomResponses []dto.RoomResponse
	userIDStr := strconv.Itoa(userID)

	for _, room := range allRooms {
		isOwner := room.OwnerId == userIDStr

		roomResponses = append(
			roomResponses, dto.RoomResponse{
				ID:      room.ID,
				Name:    room.Name,
				Avatar:  utils.ConvertToURL(room.Avatar, baseURL),
				OwnerID: room.OwnerId,
				IsOwner: isOwner,
			},
		)
	}

	return roomResponses, nil
}

func (rs *RoomService) GetRoomByID(roomID string, userID int, baseURL string) (*dto.RoomResponse, error) {
	room, err := rs.roomRepo.GetRoomById(roomID)
	if err != nil {
		return nil, err
	}

	userIDStr := strconv.Itoa(userID)
	isOwner := room.OwnerId == userIDStr

	response := &dto.RoomResponse{
		ID:      room.ID,
		Name:    room.Name,
		Avatar:  utils.ConvertToURL(room.Avatar, baseURL),
		OwnerID: room.OwnerId,
		IsOwner: isOwner,
	}

	return response, nil
}

func (rs *RoomService) GetAllRoomsWithStatus(userID int, baseURL string) ([]dto.RoomResponse, error) {
	activeRoomIDs := ActiveConnections.GetActiveRoomIDs()
	if len(activeRoomIDs) == 0 {
		return []dto.RoomResponse{}, nil
	}

	allRooms, err := rs.roomRepo.GetRoomsByIDs(activeRoomIDs)
	if err != nil {
		return nil, err
	}

	var roomResponses []dto.RoomResponse
	userIDStr := strconv.Itoa(userID)

	for _, room := range allRooms {
		isOwner := room.OwnerId == userIDStr

		roomResponses = append(
			roomResponses, dto.RoomResponse{
				ID:      room.ID,
				Name:    room.Name,
				Avatar:  utils.ConvertToURL(room.Avatar, baseURL),
				OwnerID: room.OwnerId,
				IsOwner: isOwner,
			},
		)
	}

	return roomResponses, nil
}
