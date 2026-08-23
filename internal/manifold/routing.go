package manifold

import (
	"sort"

	"github.com/lacsar712/chillrack/internal/model"
)

type Router struct{ routes []model.ManifoldRoute }

func NewRouter(routes []model.ManifoldRoute) *Router {
	cp := append([]model.ManifoldRoute(nil), routes...)
	sort.Slice(cp, func(i, j int) bool {
		if cp[i].Priority == cp[j].Priority {
			return cp[i].From < cp[j].From
		}
		return cp[i].Priority > cp[j].Priority
	})
	return &Router{routes: cp}
}

func (r *Router) Path(from, to model.ManifoldID) ([]model.ManifoldRoute, bool) {
	for _, route := range r.routes {
		if route.From == from && route.To == to {
			return []model.ManifoldRoute{route}, true
		}
	}
	return nil, false
}

func (r *Router) ValvesFor(path []model.ManifoldRoute) []model.ValveID {
	out := make([]model.ValveID, 0, len(path))
	for _, p := range path {
		out = append(out, p.Valve)
	}
	return out
}

func (r *Router) Reachable(from model.ManifoldID) []model.ManifoldID {
	seen := make(map[model.ManifoldID]struct{})
	for _, route := range r.routes {
		if route.From == from {
			seen[route.To] = struct{}{}
		}
	}
	out := make([]model.ManifoldID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}