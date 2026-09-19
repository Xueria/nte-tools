package pool

import (
	"maps"
	"math"
	"sort"
)

// PlanStep describes how a single draw is paid.
type PlanStep struct {
	Draw         int
	ResourceID   string
	ResourceName string
	Cost         int
	Remaining    int // remaining amount of the spent resource after this draw
}

// PlanResult is the outcome of calculating a spending plan.
type PlanResult struct {
	Steps        []PlanStep
	Final        map[string]int // final balance for every resource
	Insufficient bool
	FailAtDraw   int
}

// CalculatePlan computes the optimal spending plan for the inclusive draw
// range [start, end], using the following lexicographic objective:
//
//  1. maximise the number of completed draws;
//  2. minimise spending of the preferred resource (when preferredID != "");
//  3. minimise the imbalance between the remaining fractions of each resource.
//
// Because the number of draws is small, it enumerates every resource assignment
// (each draw independently chooses one of its accepted resources) and keeps
// the best one, which is guaranteed to be optimal.
func CalculatePlan(p Pool, start, end int, balances map[string]int, preferredID string) PlanResult {
	costs := buildCostLookup(p)
	order := resourceOrder(p)

	// Effective draws: only those with at least one payment option. A draw with
	// no option (missing/empty price) cannot be paid, so the plan stops before it.
	var draws []int
	unpayableDraw := 0
	for d := start; d <= end; d++ {
		if len(costs[d]) == 0 {
			unpayableDraw = d
			break
		}
		draws = append(draws, d)
	}

	res := PlanResult{Final: make(map[string]int, len(balances))}
	maps.Copy(res.Final, balances)

	if len(draws) == 0 {
		res.Insufficient = true
		res.FailAtDraw = start
		return res
	}

	best := candidate{completed: -1}
	bestChoices := make([]string, len(draws))
	choices := make([]string, len(draws))

	var search func(idx int)
	search = func(idx int) {
		if idx == len(draws) {
			c := evaluateCandidate(draws, choices, costs, balances, preferredID)
			if c.better(best) {
				best = c
				copy(bestChoices, choices)
			}
			return
		}
		for _, id := range drawResources(draws[idx], costs, order) {
			choices[idx] = id
			search(idx + 1)
		}
	}
	search(0)

	// Rebuild the result from the best assignment.
	for i := 0; i < best.completed && i < len(draws); i++ {
		id := bestChoices[i]
		cost := costs[draws[i]][id]
		res.Final[id] -= cost
		res.Steps = append(res.Steps, PlanStep{
			Draw:         draws[i],
			ResourceID:   id,
			ResourceName: ResourceName(p, id),
			Cost:         cost,
			Remaining:    res.Final[id],
		})
	}

	if best.completed < len(draws) {
		res.Insufficient = true
		res.FailAtDraw = draws[best.completed]
	} else if unpayableDraw != 0 {
		res.Insufficient = true
		res.FailAtDraw = unpayableDraw
	}

	return res
}

// candidate is the score of a single resource assignment.
type candidate struct {
	completed      int
	preferredSpent int
	imbalance      float64
}

// better reports whether c is a strictly better candidate than o, following the
// objective order: more draws, then less preferred spend, then better balance.
func (c candidate) better(o candidate) bool {
	if c.completed != o.completed {
		return c.completed > o.completed
	}
	if c.preferredSpent != o.preferredSpent {
		return c.preferredSpent < o.preferredSpent
	}
	return c.imbalance < o.imbalance
}

// evaluateCandidate walks an assignment (choices[i] pays draws[i]) and returns
// how many draws it completes before running out, plus its score.
func evaluateCandidate(draws []int, choices []string, costs map[int]map[string]int, balances map[string]int, preferredID string) candidate {
	spent := make(map[string]int, len(balances))
	completed := 0

	for i, id := range choices {
		cost := costs[draws[i]][id]
		if spent[id]+cost > balances[id] {
			break
		}
		spent[id] += cost
		completed++
	}

	preferredSpent := 0
	if preferredID != "" {
		preferredSpent = spent[preferredID]
	}

	return candidate{
		completed:      completed,
		preferredSpent: preferredSpent,
		imbalance:      imbalanceOf(balances, spent),
	}
}

// imbalanceOf measures how unbalanced the leftover amounts are across all
// resources, as a value in [0, 1] (0 = all resources left with the same
// fraction of their original balance).
func imbalanceOf(balances, spent map[string]int) float64 {
	minFrac := math.MaxFloat64
	maxFrac := -math.MaxFloat64

	for id, bal := range balances {
		if bal <= 0 {
			continue
		}
		frac := float64(bal-spent[id]) / float64(bal)
		if frac < minFrac {
			minFrac = frac
		}
		if frac > maxFrac {
			maxFrac = frac
		}
	}

	if minFrac == math.MaxFloat64 {
		return 0
	}
	return maxFrac - minFrac
}

// buildCostLookup maps a draw number (1 based) to its cost table.
func buildCostLookup(p Pool) map[int]map[string]int {
	costs := make(map[int]map[string]int, len(p.Manifest.Prices))
	for _, price := range p.Manifest.Prices {
		table := make(map[string]int, len(price.Cost))
		maps.Copy(table, price.Cost)
		costs[price.Draw] = table
	}
	return costs
}

// resourceOrder returns every resource id in a deterministic order: the
// declared pool resources first, then any cost-table keys that are not
// declared resources.
func resourceOrder(p Pool) []string {
	seen := make(map[string]bool, len(p.Resources)+1)
	order := make([]string, 0, len(p.Resources)+1)

	for _, resource := range p.Resources {
		order = append(order, resource.ID)
		seen[resource.ID] = true
	}
	for _, price := range p.Manifest.Prices {
		for id := range price.Cost {
			if !seen[id] {
				order = append(order, id)
				seen[id] = true
			}
		}
	}
	return order
}

// drawResources returns the resources accepted by a draw, in the given order.
func drawResources(draw int, costs map[int]map[string]int, order []string) []string {
	table := costs[draw]
	result := make([]string, 0, len(table))
	for _, id := range order {
		if _, ok := table[id]; ok {
			result = append(result, id)
		}
	}
	return result
}

// ResourceName resolves a resource ID to its display name.
func ResourceName(p Pool, id string) string {
	for _, resource := range p.Resources {
		if resource.ID == id {
			return resource.Name
		}
	}
	return id
}

// SortedResourceIDs returns the pool resource IDs in a stable order, used to
// present final balances consistently.
func SortedResourceIDs(p Pool) []string {
	ids := make([]string, 0, len(p.Resources))
	for _, resource := range p.Resources {
		ids = append(ids, resource.ID)
	}
	sort.Strings(ids)
	return ids
}
