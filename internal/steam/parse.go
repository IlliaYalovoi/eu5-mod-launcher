package steam

import (
	"fmt"

	"github.com/tidwall/gjson"
)

var errPublishedDetailsMissing = fmt.Errorf("missing publishedfiledetails array")

func parseWorkshopResponse(body []byte, requestedIDs map[string]struct{}) (map[string]WorkshopItem, error) {
	details := gjson.GetBytes(body, "response.publishedfiledetails")
	if !details.Exists() {
		return nil, errPublishedDetailsMissing
	}

	items := make(map[string]WorkshopItem)
	details.ForEach(func(_, detail gjson.Result) bool {
		itemID := detail.Get("publishedfileid").Str
		if itemID == "" {
			return true
		}
		if len(requestedIDs) > 0 {
			if _, wanted := requestedIDs[itemID]; !wanted {
				return true
			}
		}
		if result := detail.Get("result").Int(); result != 1 {
			return true
		}

		items[itemID] = WorkshopItem{
			ItemID:      itemID,
			Title:       detail.Get("title").Str,
			Description: detail.Get("description").Str,
			PreviewURL:  detail.Get("preview_url").Str,
		}
		return true
	})

	return items, nil
}
