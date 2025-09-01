package dto

type CreateRoomRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type UpdateRoomRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type RoomResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	OwnerID string `json:"owner_id"`
	IsOwner bool   `json:"is_owner"`
}

type RoomListResponse struct {
	MyRooms    []RoomResponse `json:"my_rooms"`
	OtherRooms []RoomResponse `json:"other_rooms"`
}

type RoomMembersResponse struct {
	Members       []RoomMemberInfo `json:"members"`
	OnlineMembers []int            `json:"online_members"`
}

type RoomMemberInfo struct {
	UserID    int    `json:"user_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Avatar    string `json:"avatar"`
	IsOnline  bool   `json:"is_online"`
}
