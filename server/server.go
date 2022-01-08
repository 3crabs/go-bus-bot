package server

import "github.com/3crabs/go-bus-bot/user"

type server struct {
	users map[int64]*user.User
}

func NewServer() *server {
	return &server{users: make(map[int64]*user.User)}
}

func (s *server) GetUser(chatId int64) *user.User {
	_, ok := s.users[chatId]
	if !ok {
		s.users[chatId] = user.NewUser()
	}
	u, _ := s.users[chatId]
	return u
}
