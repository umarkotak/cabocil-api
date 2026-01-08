package book_link_service

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/umarkotak/ytkidd-api/contract_resp"
	"github.com/umarkotak/ytkidd-api/repos/book_link_repo"
	"github.com/umarkotak/ytkidd-api/repos/user_subscription_repo"
)

func GetAllGrouped(ctx context.Context, userGuid string) ([]contract_resp.BookLinkGroup, error) {
	bookLinks, err := book_link_repo.GetAll(ctx)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
		return nil, err
	}

	// Group book links by group_name
	groupMap := make(map[string][]contract_resp.BookLink)
	groupOrder := []string{}

	isSubscribed, err := user_subscription_repo.IsUserGuidHasActiveSubscription(ctx, userGuid)
	if err != nil {
		logrus.WithContext(ctx).Error(err)
	}

	for _, bl := range bookLinks {
		url := bl.Url
		if bl.Premium && !isSubscribed {
			url = ""
		}

		bookLinkResp := contract_resp.BookLink{
			Name:     bl.Name,
			Url:      url,
			ImageUrl: bl.ImageUrl,
			Premium:  bl.Premium,
		}

		if _, exists := groupMap[bl.GroupName]; !exists {
			groupOrder = append(groupOrder, bl.GroupName)
		}
		groupMap[bl.GroupName] = append(groupMap[bl.GroupName], bookLinkResp)
	}

	// Build response maintaining insertion order
	result := make([]contract_resp.BookLinkGroup, 0, len(groupOrder))
	for _, groupName := range groupOrder {
		result = append(result, contract_resp.BookLinkGroup{
			GroupName: groupName,
			BookLinks: groupMap[groupName],
		})
	}

	return result, nil
}
