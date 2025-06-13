package plan

import (
	"fmt"
	"strings"
	"sync"

	tfJson "github.com/hashicorp/terraform-json"
	"github.com/orange-car/tfplan/internal/helpers"
	"github.com/vodkaslime/wildcard"
)

type Diff struct {
	// The value of the attribute before the planned change.
	Before string `json:"before"`
	// The value of the attribute after the planned change.
	After string `json:"after"`
}

type Filter struct {
	// A wildcard-supported string to match against entity addresses.
	NamePattern string `json:"namePattern"`
	// A wildcard-supported map string of slice Diff to match against
	// the entity planned change. The key is the field within the resource/
	// output/drift.
	DiffPatterns map[string][]Diff `json:"diffPatterns"`
}

type InspectFilter struct {
	// Filter criteria to exclude (filter) output changes.
	OutputChanges []Filter `json:"outputChanges"`
	// Filter criteria to exclude (filter) resource changes.
	ResourceChanges []Filter `json:"resourceChanges"`
	// Filter criteria to exclude (filter) drift resource changes.
	DriftChanges []Filter `json:"driftChanges"`
}

type InspectInput struct {
	// Optional filter to apply to the plan during inspection.
	Filter *InspectFilter `json:"filter"`
}

// Differences in attributes between two entities. Map key is the attribute. Map
// value is the difference in that attribute
type EntityDiff map[string]*Diff

// The identified diffs within a plan
type InspectDiff struct {
	// Planned changes to resources.
	Resources map[string]EntityDiff `json:"resources"`
	// Planned changes to outputs.
	Outputs map[string]EntityDiff `json:"outputs"`
	// Planned changes to resource drifts.
	ResourceDrifts map[string]EntityDiff `json:"resourceDrifts"`
}

// Result of calling Inspect() to inspect a Terraform plan.
type InspectOutput struct {
	// The identified diffs within a plan
	Diff *InspectDiff `json:"diff"`
}

/*
Recursively traverses any object and transforms it into a
map string of string where the key is the jmespath-style address
and the value is the value.
*/
func flatten(path string, a any, kvPairs map[string]string) {

	switch av1 := a.(type) {
	case map[string]any:
		path = path + "."

		for k, v := range av1 {
			flatten(path+k, v, kvPairs)
		}

	case []any:
		path = path + "."

		for i, elem := range av1 {
			flatten(fmt.Sprintf("%s[%v]", path, i), elem, kvPairs)
		}

	case []string:
		path = path + "."
		strs := []string{}
		for _, v := range av1 {
			strs = append(strs, fmt.Sprintf("%v", v))
		}
		kvPairs[path] = strings.Join(strs, ",")

	default:
		if path == "" {
			path = path + "."
		}

		if a == nil {
			return
		}
		kvPairs[path] = fmt.Sprintf("%v", a)
	}
}

/*
takes a change of any kind and converts it into an EntityDiff.
*/
func parseChange(chng *tfJson.Change) EntityDiff {
	out := EntityDiff{}

	afterUnknowns := map[string]string{}
	beforeSensitives := map[string]string{}
	afterSensitives := map[string]string{}
	before := map[string]string{}
	after := map[string]string{}

	wg := sync.WaitGroup{}

	wg.Add(5)
	go func() {
		if chng.AfterUnknown != false {
			flatten("", chng.AfterUnknown, afterUnknowns)
		}

		wg.Done()
	}()
	go func() {
		if chng.BeforeSensitive != false {
			flatten("", chng.BeforeSensitive, beforeSensitives)
		}
		wg.Done()
	}()
	go func() {
		if chng.AfterSensitive != false {
			flatten("", chng.AfterSensitive, afterSensitives)
		}
		wg.Done()
	}()
	go func() {
		flatten("", chng.Before, before)
		wg.Done()
	}()
	go func() {
		flatten("", chng.After, after)
		wg.Done()
	}()
	wg.Wait()

	// Must do unknowns before sensitives. Desired behaviour is
	// sensitives that are also unknown are marked as sensitive.
	for k, b := range afterUnknowns {
		if b == "true" {
			after[k] = "(known after apply)"
		}
	}

	for k, b := range beforeSensitives {
		if b == "true" {
			before[k] = "(sensitive value)"
		}
	}

	for k, b := range afterSensitives {
		if b == "true" {
			after[k] = "(sensitive value)"
		}
	}

	for path, beforeVal := range before {
		afterVal, ok := after[path]

		if !ok {
			out[path] = &Diff{
				Before: beforeVal,
				After:  "(empty)",
			}
			continue
		}

		if afterVal != beforeVal {
			out[path] = &Diff{
				Before: beforeVal,
				After:  afterVal,
			}
			continue
		}
	}

	for path, afterVal := range after {

		_, ok := before[path]
		if !ok {
			out[path] = &Diff{
				Before: "(empty)",
				After:  afterVal,
			}
			continue
		}
	}

	return out
}

/*
Checks if a InspectOutput is empty
*/
func (i *InspectOutput) IsEmpty() bool {
	return len(i.Diff.Outputs) == 0 && len(i.Diff.ResourceDrifts) == 0 && len(i.Diff.Resources) == 0
}

func filterEntityDiffs(address string, entityDiff EntityDiff, filters []Filter, inspectDiffMap map[string]EntityDiff) (map[string]EntityDiff, error) {
	m := wildcard.NewMatcher()

	for _, filter := range filters {
		if match, err := m.Match(filter.NamePattern, address); err != nil {
			return nil, fmt.Errorf("unable to match %s with pattern %s caused by: %v", address, filter.NamePattern, err)
		} else if match {
			// The name has matched a filter rule. Now to check if any of the diff patterns apply to the entity's diffs

			for path, diff := range entityDiff {
				for pathPattern, diffPatterns := range filter.DiffPatterns {
					if match, err := m.Match(pathPattern, path); err != nil {
						return nil, fmt.Errorf("unable to match %s.%s with pattern %s caused by: %v", address, path, pathPattern, err)
					} else if match {
						// The path has matched a filter rule. Now to check if the before and after patterns apply

						for _, diffPattern := range diffPatterns {
							bMatch, err := m.Match(diffPattern.Before, diff.Before)
							if err != nil {
								return nil, fmt.Errorf("unable to match %s.%s before value %s with pattern %s caused by: %v", address, path, diff.Before, diffPattern.Before, err)
							}
							aMatch, err := m.Match(diffPattern.After, diff.After)
							if err != nil {
								return nil, fmt.Errorf("unable to match %s.%s after value %s with pattern %s caused by: %v", address, path, diff.After, diffPattern.After, err)
							}

							if bMatch && aMatch {
								// The before and after patterns both match. This diff should be filtered out

								delete(inspectDiffMap[address], path)
								if len(inspectDiffMap[address]) == 0 {
									// There are no more diffs for this entity. Remove it from the output

									delete(inspectDiffMap, address)
								}
							}
						}
					}
				}
			}
		}
	}
	return inspectDiffMap, nil
}

func (i *InspectFilter) apply(in *InspectDiff) (*InspectDiff, error) {
	// m := wildcard.NewMatcher()

	inspectDiff := in

	for address, entDiff := range in.Resources {
		var err error
		in.Resources, err = filterEntityDiffs(address, entDiff, i.ResourceChanges, in.Resources)

		if err != nil {
			return in, fmt.Errorf("unable to apply filters to resource at address %s caused by: %v", address, err)
		}
	}

	for address, entDiff := range in.ResourceDrifts {
		var err error
		in.ResourceDrifts, err = filterEntityDiffs(address, entDiff, i.DriftChanges, in.ResourceDrifts)

		if err != nil {
			return in, fmt.Errorf("unable to apply filters to resource drift at address %s caused by: %v", address, err)
		}
	}

	for name, entDiff := range in.Outputs {
		var err error
		in.Outputs, err = filterEntityDiffs(name, entDiff, i.OutputChanges, in.Outputs)

		if err != nil {
			return in, fmt.Errorf("unable to apply filters to output name %s caused by: %v", name, err)
		}
	}

	return inspectDiff, nil
}

/*
Inspects the plan with the provided filter to find changes not
captured by the filter.
*/
func (p *Plan) Inspect(params *InspectInput) (*InspectOutput, error) {

	out := &InspectOutput{
		Diff: &InspectDiff{
			Resources:      map[string]EntityDiff{},
			ResourceDrifts: map[string]EntityDiff{},
			Outputs:        map[string]EntityDiff{},
		},
	}

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		for _, rChange := range p.ResourceChanges {
			out.Diff.Resources[rChange.Address] = parseChange(rChange.Change)
		}
		wg.Done()
	}()

	go func() {
		for _, dChange := range p.ResourceDrift {
			out.Diff.ResourceDrifts[dChange.Address] = parseChange(dChange.Change)
		}
		wg.Done()
	}()

	go func() {
		for name, oChange := range p.OutputChanges {
			out.Diff.Outputs[name] = parseChange(oChange)
		}
		wg.Done()
	}()

	wg.Wait()

	var err error
	out.Diff, err = params.Filter.apply(out.Diff)
	if err != nil {
		return nil, fmt.Errorf("failed to apply filter caused by: %e", err)
	}

	return out, nil
}

const (
	colorNone   = "\033[0m"
	colorBold   = "\033[1m"
	colorOrange = "\033[33m"
)

/*
Produces a slice of strings output which can be printed line by line
to get a Terraform-style stdout report of the inspect.
*/
func (o *InspectOutput) Pretty() []string {
	var out []string
	out = append(out, "\tTerraform plan contained the following un-filtered changes:\n")

	for address, diffs := range o.Diff.Resources {
		out = append(out, fmt.Sprintf("\n\t\tresource %s\"%s\"%s changes:\n", colorBold, address, colorNone))
		maxWidth := 0
		for path := range diffs {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range diffs {
			out = append(out, fmt.Sprintf("\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	for address, diffs := range o.Diff.ResourceDrifts {
		out = append(out, fmt.Sprintf("\n\t\tresource %s\"%s\"%s drift:\n", colorBold, address, colorNone))
		maxWidth := 0
		for path := range diffs {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range diffs {
			out = append(out, fmt.Sprintf("\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	for name, diffs := range o.Diff.Outputs {
		out = append(out, fmt.Sprintf("\n\t\toutput %s\"%s\"%s changes:\n", colorBold, name, colorNone))
		maxWidth := 0
		for path := range diffs {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range diffs {
			out = append(out, fmt.Sprintf("\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	out = append(out, fmt.Sprintf("\n\tChanges: %v resources, %v resource drifts, %v outputs\n", len(o.Diff.Resources), len(o.Diff.ResourceDrifts), len(o.Diff.Outputs)))
	return out
}
