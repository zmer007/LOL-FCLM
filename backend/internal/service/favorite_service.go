package service

import (
	"errors"

	"roadside-league/backend/internal/model"
	"roadside-league/backend/internal/repository"
)

var ErrFavoriteSelf = errors.New("cannot favorite own content")

// FavoriteService 收藏业务。
type FavoriteService struct {
	favorites *repository.FavoriteRepository
	contents  *ContentService
}

func NewFavoriteService(fav *repository.FavoriteRepository, contents *ContentService) *FavoriteService {
	return &FavoriteService{favorites: fav, contents: contents}
}

// Toggle 切换收藏状态，返回切换后是否已收藏。
func (s *FavoriteService) Toggle(userID, contentID int64) (bool, error) {
	exists, err := s.favorites.Exists(userID, contentID)
	if err != nil {
		return false, err
	}
	if exists {
		if err := s.favorites.Delete(userID, contentID); err != nil {
			return false, err
		}
		return false, nil
	}
	if err := s.favorites.Create(&model.Favorite{UserID: userID, ContentID: contentID}); err != nil {
		return false, err
	}
	return true, nil
}

// ListMine 返回当前用户收藏的内容列表（按收藏时间倒序）。
func (s *FavoriteService) ListMine(userID int64) ([]ContentVO, error) {
	favs, err := s.favorites.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.ContentID)
	}
	return s.contents.ListByIDs(ids, userID)
}
