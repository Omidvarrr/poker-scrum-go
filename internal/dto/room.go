package dto

type CreateRoomRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type UpdateRoomRequest struct {
	Name   string `json:"name" form:"name"`
	Avatar string `json:"avatar" form:"avatar"`
}

type JoinRoomRequest struct {
	RoomID string `json:"room_id"`
}

type HandleJoinRequestRequest struct {
	JoinRequestID int    `json:"join_request_id"`
	Action        string `json:"action"` // approve or reject
}

type RoomResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	OwnerID  string `json:"owner_id"`
	IsOwner  bool   `json:"is_owner"`
	IsMember bool   `json:"is_member"`
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
	Role      string `json:"role"`
	IsOnline  bool   `json:"is_online"`
}

type ManageRoomMemberRequest struct {
	UserID int    `json:"user_id"`
	Action string `json:"action"` // promote_to_admin, revoke_admin, remove
}
