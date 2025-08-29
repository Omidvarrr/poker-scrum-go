package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/models"
	"awesomeProject1/internal/repositories"
	"awesomeProject1/internal/utils"
	"errors"
	"github.com/google/uuid"
	"strconv"
)

type RoomService struct {
	roomRepo        repositories.RoomRepository
	joinRequestRepo repositories.JoinRequestRepository
	userRepo        repositories.UserRepository
}

func NewRoomService(roomRepo repositories.RoomRepository, joinRequestRepo repositories.JoinRequestRepository, userRepo repositories.UserRepository) *RoomService {
	return &RoomService{
		roomRepo:        roomRepo,
		joinRequestRepo: joinRequestRepo,
		userRepo:        userRepo,
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

	err = rs.roomRepo.AddRoomMember(roomID, userID, models.RoleAdmin)
	if err != nil {
		return dto.RoomResponse{}, err
	}

	return dto.RoomResponse{
		ID:       createdRoom.ID,
		Name:     createdRoom.Name,
		Avatar:   createdRoom.Avatar,
		OwnerID:  createdRoom.OwnerId,
		IsOwner:  true,
		IsMember: true,
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
		myRoomResponses = append(myRoomResponses, dto.RoomResponse{
			ID:       room.ID,
			Name:     room.Name,
			Avatar:   room.Avatar,
			OwnerID:  room.OwnerId,
			IsOwner:  true,
			IsMember: true,
		})
	}

	for _, room := range allRooms {
		if room.OwnerId != userIDStr {
			isMember, _ := rs.roomRepo.IsUserInRoom(room.ID, userID)
			otherRoomResponses = append(otherRoomResponses, dto.RoomResponse{
				ID:       room.ID,
				Name:     room.Name,
				Avatar:   room.Avatar,
				OwnerID:  room.OwnerId,
				IsOwner:  false,
				IsMember: isMember,
			})
		}
	}

	return dto.RoomListResponse{
		MyRooms:    myRoomResponses,
		OtherRooms: otherRoomResponses,
	}, nil
}

func (rs *RoomService) RequestJoinRoom(userID int, roomID string) error {
	_, err := rs.roomRepo.GetRoomById(roomID)
	if err != nil {
		return err
	}

	isMember, err := rs.roomRepo.IsUserInRoom(roomID, userID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a member of this room")
	}

	hasPending, err := rs.joinRequestRepo.HasPendingRequest(roomID, userID)
	if err != nil {
		return err
	}
	if hasPending {
		return errors.New("user already has a pending request for this room")
	}

	request := models.JoinRequest{
		RoomID: roomID,
		UserID: userID,
		Status: models.JoinRequestStatusPending,
	}

	_, err = rs.joinRequestRepo.CreateJoinRequest(request)
	return err
}

func (rs *RoomService) GetJoinRequests(userID int, roomID string) ([]models.JoinRequest, error) {
	isOwnerOrAdmin, err := rs.isOwnerOrAdmin(userID, roomID)
	if err != nil {
		return nil, err
	}
	if !isOwnerOrAdmin {
		return nil, errors.New("user is not authorized to view join requests")
	}

	return rs.joinRequestRepo.GetJoinRequestsByRoom(roomID)
}

func (rs *RoomService) HandleJoinRequest(userID int, requestID int, action string) error {
	request, err := rs.joinRequestRepo.GetJoinRequestById(requestID)
	if err != nil {
		return err
	}

	isOwnerOrAdmin, err := rs.isOwnerOrAdmin(userID, request.RoomID)
	if err != nil {
		return err
	}
	if !isOwnerOrAdmin {
		return errors.New("user is not authorized to handle join requests")
	}

	var status models.JoinRequestStatus
	switch action {
	case "approve":
		status = models.JoinRequestStatusApproved
		err = rs.roomRepo.AddRoomMember(request.RoomID, request.UserID, models.RoleMember)
		if err != nil {
			return err
		}
	case "reject":
		status = models.JoinRequestStatusRejected
	default:
		return errors.New("invalid action")
	}

	return rs.joinRequestRepo.UpdateJoinRequestStatus(requestID, status, userID)
}

func (rs *RoomService) GetUserJoinRequests(userID int) ([]models.JoinRequest, error) {
	return rs.joinRequestRepo.GetJoinRequestsByUser(userID)
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

func (rs *RoomService) ManageRoomMember(userID int, roomID string, request dto.ManageRoomMemberRequest) error {
	isOwner, err := rs.roomRepo.IsUserOwner(roomID, userID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only room owner can manage members")
	}

	switch request.Action {
	case "promote_to_admin":
		return rs.roomRepo.UpdateMemberRole(roomID, request.UserID, models.RoleAdmin)
	case "revoke_admin":
		return rs.roomRepo.UpdateMemberRole(roomID, request.UserID, models.RoleMember)
	case "remove":
		return rs.roomRepo.RemoveRoomMember(roomID, request.UserID)
	default:
		return errors.New("invalid action")
	}
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

func (rs *RoomService) GetRoomMembers(userID int, roomID string) (dto.RoomMembersResponse, error) {
	isMember, err := rs.roomRepo.IsUserInRoom(roomID, userID)
	if err != nil {
		return dto.RoomMembersResponse{}, err
	}
	if !isMember {
		return dto.RoomMembersResponse{}, errors.New("user is not a member of this room")
	}

	members, err := rs.roomRepo.GetRoomMembers(roomID)
	if err != nil {
		return dto.RoomMembersResponse{}, err
	}

	var memberInfos []dto.RoomMemberInfo
	for _, member := range members {
		user, err := rs.userRepo.GetUserById(member.UserID)
		if err != nil {
			continue
		}

		memberInfos = append(memberInfos, dto.RoomMemberInfo{
			UserID:   user.ID,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Role:     string(member.Role),
			IsOnline: false,
		})
	}

	return dto.RoomMembersResponse{
		Members:       memberInfos,
		OnlineMembers: []int{},
	}, nil
}

func (rs *RoomService) isOwnerOrAdmin(userID int, roomID string) (bool, error) {
	isOwner, err := rs.roomRepo.IsUserOwner(roomID, userID)
	if err != nil {
		return false, err
	}
	if isOwner {
		return true, nil
	}

	return rs.roomRepo.IsUserAdmin(roomID, userID)
}

func (rs *RoomService) GetJoinedRooms(userID int, baseURL string) ([]dto.RoomResponse, error) {
	rooms, err := rs.roomRepo.GetRoomsByUser(userID)
	if err != nil {
		return nil, err
	}
	
	var roomResponses []dto.RoomResponse
	for _, room := range rooms {
		userIDStr := strconv.Itoa(userID)
		isOwner := room.OwnerId == userIDStr
		
		roomResponses = append(roomResponses, dto.RoomResponse{
			ID:       room.ID,
			Name:     room.Name,
			Avatar:   utils.ConvertToURL(room.Avatar, baseURL),
			OwnerID:  room.OwnerId,
			IsOwner:  isOwner,
			IsMember: true,
		})
	}
	
	return roomResponses, nil
}

func (rs *RoomService) GetAllRoomsWithStatus(userID int, baseURL string) ([]dto.RoomResponse, error) {
	allRooms, err := rs.roomRepo.GetAllRooms()
	if err != nil {
		return nil, err
	}
	
	var roomResponses []dto.RoomResponse
	userIDStr := strconv.Itoa(userID)
	
	for _, room := range allRooms {
		isOwner := room.OwnerId == userIDStr
		isMember, _ := rs.roomRepo.IsUserInRoom(room.ID, userID)
		
		roomResponses = append(roomResponses, dto.RoomResponse{
			ID:       room.ID,
			Name:     room.Name,
			Avatar:   utils.ConvertToURL(room.Avatar, baseURL),
			OwnerID:  room.OwnerId,
			IsOwner:  isOwner,
			IsMember: isMember,
		})
	}
	
	return roomResponses, nil
}
