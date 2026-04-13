package steam

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	errSteamResponseObjectMissing = errors.New("missing response object")
	errPublishedDetailsMissing    = errors.New("missing publishedfiledetails array")
	errSteamNodeKeyMissing        = errors.New("steam response node key missing")
	errSteamNodeUnexpectedType    = errors.New("steam response node has unexpected type")
)

func parseWorkshopResponse(body []byte, requestedIDs map[string]struct{}) (map[string]WorkshopItem, error) {
	var root map[string]any
	if err := json.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("decode steam response json: %w", err)
	}

	responseMap, err := mapNode(root, "response")
	if err != nil {
		return nil, err
	}
	detailsList, err := listNode(responseMap, "publishedfiledetails")
	if err != nil {
		return nil, err
	}

	items := make(map[string]WorkshopItem, len(detailsList))
	for i := range detailsList {
		detailMap, ok := detailsList[i].(map[string]any)
		if !ok {
			continue
		}

		itemID := stringValue(detailMap, "publishedfileid")
		if itemID == "" {
			continue
		}
		if len(requestedIDs) > 0 {
			if _, wanted := requestedIDs[itemID]; !wanted {
				continue
			}
		}
		if resultCode, ok := asFloat(detailMap["result"]); ok && int(resultCode) != 1 {
			continue
		}

		items[itemID] = WorkshopItem{
			ItemID:      itemID,
			Title:       stringValue(detailMap, "title"),
			Description: stringValue(detailMap, "description"),
			PreviewURL:  stringValue(detailMap, "preview_url"),
		}
	}

	return items, nil
}

func mapNode(root map[string]any, key string) (map[string]any, error) {
	node, ok := root[key]
	if !ok {
		if key == "response" {
			return nil, fmt.Errorf("%w", errSteamResponseObjectMissing)
		}
		return nil, fmt.Errorf("%w: key %q", errSteamNodeKeyMissing, key)
	}
	mapped, ok := node.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: key %q type %T", errSteamNodeUnexpectedType, key, node)
	}
	return mapped, nil
}

func listNode(root map[string]any, key string) ([]any, error) {
	node, ok := root[key]
	if !ok {
		if key == "publishedfiledetails" {
			return nil, fmt.Errorf("%w", errPublishedDetailsMissing)
		}
		return nil, fmt.Errorf("%w: key %q", errSteamNodeKeyMissing, key)
	}
	list, ok := node.([]any)
	if !ok {
		return nil, fmt.Errorf("%w: key %q type %T", errSteamNodeUnexpectedType, key, node)
	}
	return list, nil
}

func stringValue(data map[string]any, key string) string {
	value, ok := data[key]
	if !ok {
		return ""
	}
	parsed, ok := value.(string)
	if !ok {
		return ""
	}
	return parsed
}

func asFloat(value any) (float64, bool) {
	parsed, ok := value.(float64)
	return parsed, ok
}
