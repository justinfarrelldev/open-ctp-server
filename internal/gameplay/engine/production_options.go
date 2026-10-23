package engine

import (
	"errors"
	"sort"
)

// ProductionError carries a deliberately coarse public reason. Detail is private:
// prerequisite failures must not identify hidden wonders, cities or their owners.
type ProductionError struct {
	Code   string
	Detail error
}

func (e *ProductionError) Error() string { return e.Detail.Error() }
func (e *ProductionError) Unwrap() error { return e.Detail }
func productionFailure(code, detail string) error {
	return &ProductionError{Code: code, Detail: errors.New(detail)}
}

// ProductionOptions is an exact-state append query, not authorization for an
// arbitrary replacement. Whole edits revalidate all retained entries in order.
// Items is always an array; complete-empty differs from unavailable.
type ProductionOptions struct {
	Revision    int64            `json:"revision"`
	CityID      string           `json:"city_id"`
	ContentHash string           `json:"content_hash"`
	Operation   string           `json:"operation"`
	Status      string           `json:"status"`
	Items       []ProductionItem `json:"items"`
}

// ProductionOptions requires an owned city. Callers must additionally enforce
// phase/controller authorization before publishing a complete result.
func (i *Inspector) ProductionOptions(cityID, actor string, authorized bool) *ProductionOptions {
	city, ok := cityByID(i.state, cityID)
	if !ok || city.OwnerID != actor {
		return nil
	}
	result := &ProductionOptions{Revision: i.state.Revision, CityID: cityID, ContentHash: i.state.Rules.ContentHash, Operation: "append", Status: "unavailable", Items: []ProductionItem{}}
	if !authorized || i.runtime == nil {
		return result
	}
	// Validate existing entries once, without allocating identities or cloning the
	// world for each possible item. The prefix preserves ordered duplicate checks.
	prefix := *city
	for index, retained := range city.Queue {
		prefix.Queue = city.Queue[:index]
		retained.EntryID = ""
		if _, err := validateProductionAppend(i.state, actor, prefix, retained, i.runtime); err != nil {
			return result
		}
	}
	result.Status = "complete"
	if len(city.Queue) >= MaxProductionQueueEntries {
		return result
	}
	ids := make(map[string]bool, len(i.runtime.Units)+len(i.runtime.Production))
	for _, item := range i.runtime.Units {
		ids[item.ID] = true
	}
	for _, item := range i.runtime.Production {
		ids[item.ID] = true
	}
	ordered := make([]string, 0, len(ids))
	for id := range ids {
		ordered = append(ordered, id)
	}
	sort.Strings(ordered)
	for _, id := range ordered {
		item, err := canonicalProductionItem(i.runtime, id)
		if err != nil {
			continue
		}
		item, err = validateProductionAppend(i.state, actor, *city, item, i.runtime)
		if err == nil {
			result.Items = append(result.Items, item)
		}
	}
	return result
}
