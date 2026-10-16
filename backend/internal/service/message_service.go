package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

const (
	messageMaxLen  = 500
	messageRateGap = 10 * time.Second
)

var (
	ErrMessageEmpty       = errors.New("message content is empty")
	ErrMessageTooLong     = errors.New("message content too long")
	ErrMessageRateLimited = errors.New("message rate limited")
)

type MessageService struct {
	msgs  *repository.MessageRepository
	users *repository.UserRepository
	valkey *redis.Client
}

func NewMessageService(msgs *repository.MessageRepository, users *repository.UserRepository, valkey *redis.Client) *MessageService {
	return &MessageService{msgs: msgs, users: users, valkey: valkey}
}

// Send 普通用户提交留言：校验内容 + 10s 频率限制 + 入库。
func (s *MessageService) Send(userID int64, content string) (*model.Message, error) {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil, ErrMessageEmpty
	}
	if utf8.RuneCountInString(content) > messageMaxLen {
		return nil, ErrMessageTooLong
	}

	ctx := context.Background()
	rateKey := "msg:rate:" + strconv.FormatInt(userID, 10)
	ok, err := s.valkey.SetNX(ctx, rateKey, "1", messageRateGap).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrMessageRateLimited
	}

	user, err := s.users.FindByID(userID)
	if err != nil {
		_ = s.valkey.Del(ctx, rateKey).Err()
		return nil, err
	}

	msg := &model.Message{
		UserID:   userID,
		Username: user.Username,
		Content:  content,
	}
	if err := s.msgs.Create(msg); err != nil {
		_ = s.valkey.Del(ctx, rateKey).Err()
		return nil, err
	}
	return msg, nil
}

// ListAll 管理员查看所有留言。
func (s *MessageService) ListAll() ([]model.Message, error) {
	return s.msgs.ListAll()
}

// ListMine 某用户查看自己的留言。
func (s *MessageService) ListMine(userID int64) ([]model.Message, error) {
	return s.msgs.ListByUser(userID)
}
