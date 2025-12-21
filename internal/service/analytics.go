package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
)

type AnalyticsService struct {
	db *database.Queries
}

func NewAnalyticsService(db *database.Queries) *AnalyticsService {
	return &AnalyticsService{db: db}
}

type TenantStats struct {
	TotalMembers  int64 `json:"total_members"`
	TotalPosts    int64 `json:"total_posts"`
	TotalComments int64 `json:"total_comments"`
	TotalViews    int64 `json:"total_views"`
}

type GrowthPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

type TopPost struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	ViewCount    int32     `json:"view_count"`
	LikeCount    int32     `json:"like_count"`
	CommentCount int64     `json:"comment_count"`
	CreatedAt    time.Time `json:"created_at"`
}

type RecentMember struct {
	UserID      uuid.UUID `json:"user_id"`
	DisplayName string    `json:"display_name"`
	UserName    string    `json:"user_name"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
}

func (s *AnalyticsService) GetStats(ctx context.Context, tenantID uuid.UUID) (*TenantStats, error) {
	stats, err := s.db.GetTenantStats(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// TotalViews comes as interface{} from COALESCE, need to convert
	var totalViews int64
	if v, ok := stats.TotalViews.(int64); ok {
		totalViews = v
	}

	return &TenantStats{
		TotalMembers:  stats.TotalMembers,
		TotalPosts:    stats.TotalPosts,
		TotalComments: stats.TotalComments,
		TotalViews:    totalViews,
	}, nil
}

func (s *AnalyticsService) GetMembersGrowth(ctx context.Context, tenantID uuid.UUID, days int) ([]GrowthPoint, error) {
	since := time.Now().AddDate(0, 0, -days)

	rows, err := s.db.GetMembersGrowth(ctx, database.GetMembersGrowthParams{
		TenantID: tenantID,
		JoinedAt: since,
	})
	if err != nil {
		return nil, err
	}

	result := make([]GrowthPoint, len(rows))
	for i, row := range rows {
		result[i] = GrowthPoint{
			Date:  row.Date.Format("2006-01-02"),
			Count: row.Count,
		}
	}

	return result, nil
}

func (s *AnalyticsService) GetPostsGrowth(ctx context.Context, tenantID uuid.UUID, days int) ([]GrowthPoint, error) {
	since := time.Now().AddDate(0, 0, -days)

	rows, err := s.db.GetPostsGrowth(ctx, database.GetPostsGrowthParams{
		TenantID:  tenantID,
		CreatedAt: since,
	})
	if err != nil {
		return nil, err
	}

	result := make([]GrowthPoint, len(rows))
	for i, row := range rows {
		result[i] = GrowthPoint{
			Date:  row.Date.Format("2006-01-02"),
			Count: row.Count,
		}
	}

	return result, nil
}

func (s *AnalyticsService) GetTopPosts(ctx context.Context, tenantID uuid.UUID, limit int32) ([]TopPost, error) {
	rows, err := s.db.GetTopPosts(ctx, database.GetTopPostsParams{
		TenantID: tenantID,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]TopPost, len(rows))
	for i, row := range rows {
		result[i] = TopPost{
			ID:           row.ID,
			Title:        row.Title,
			ViewCount:    row.ViewCount,
			LikeCount:    row.LikeCount,
			CommentCount: row.CommentCount,
			CreatedAt:    row.CreatedAt,
		}
	}

	return result, nil
}

func (s *AnalyticsService) GetRecentMembers(ctx context.Context, tenantID uuid.UUID, limit int32) ([]RecentMember, error) {
	rows, err := s.db.GetRecentMembers(ctx, database.GetRecentMembersParams{
		TenantID: tenantID,
		Limit:    limit,
	})
	if err != nil {
		return nil, err
	}

	result := make([]RecentMember, len(rows))
	for i, row := range rows {
		displayName := row.UserName
		if row.DisplayName.Valid && row.DisplayName.String != "" {
			displayName = row.DisplayName.String
		}

		avatarURL := ""
		if row.AvatarUrl.Valid {
			avatarURL = row.AvatarUrl.String
		}

		result[i] = RecentMember{
			UserID:      row.UserID,
			DisplayName: displayName,
			UserName:    row.UserName,
			AvatarURL:   avatarURL,
			JoinedAt:    row.JoinedAt,
		}
	}

	return result, nil
}
