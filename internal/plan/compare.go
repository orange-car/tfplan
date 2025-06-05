package plan

import (
	"fmt"

	"github.com/orange-car/tfplan/internal/helpers"
)

// The comparison between two specific entities of the same
// address/name from different inspected plans. Represented as
// PlanA and PlanB.
type CompareEntityDiff struct {
	// Map of paths (field/attributes) changed for the entity in
	// inspected plan A
	PlanA EntityDiff `json:"planA"`
	// Map of paths (field/attributes) changed for the entity in
	// inspected plan B
	PlanB EntityDiff `json:"planB"`
}

// The identified diffs (divergence) between the two plan diffs
type CompareDiff struct {
	// Set of resources and their respective diverging plan a and b differences.
	Resources map[string]CompareEntityDiff `json:"resources"`
	// Set of outputs and their respective diverging plan a and b differences.
	Outputs map[string]CompareEntityDiff `json:"outputs"`
	// Set of resource drifts and their respective diverging plan a and b differences.
	ResourceDrifts map[string]CompareEntityDiff `json:"resourceDrifts"`
}

// Result of calling CompareInspects() to identify divergences of two inspected plans
type CompareInspectsOutput struct {
	// The identified diffs (divergence) between the two plan diffs
	Diff *CompareDiff `json:"diff"`
}

/*
Checks if a CompareInspectsOutput is empty
*/
func (c *CompareInspectsOutput) IsEmpty() bool {
	return len(c.Diff.Outputs) == 0 && len(c.Diff.ResourceDrifts) == 0 && len(c.Diff.ResourceDrifts) == 0
}

func getEntityDiffDivergence(a, b map[string]EntityDiff) map[string]CompareEntityDiff {
	out := map[string]CompareEntityDiff{}

	// Go over each item (A)
	for aAddress, aEntityDiff := range a {

		if _, ok := b[aAddress]; !ok {
			//a address not in b

			out[aAddress] = CompareEntityDiff{
				PlanA: aEntityDiff,
				PlanB: EntityDiff{},
			}

			for path := range aEntityDiff {
				out[aAddress].PlanB[path] = &Diff{
					Before: "(empty)",
					After:  "(empty)",
				}
			}
			continue
		}

		// a address is also in b. Go over each path
		for aPath, aDiff := range aEntityDiff {

			if _, ok := b[aAddress][aPath]; !ok {
				// a path for address is not in b address

				if _, exists := out[aAddress]; !exists {
					out[aAddress] = CompareEntityDiff{
						PlanA: EntityDiff{},
						PlanB: EntityDiff{},
					}
				}

				out[aAddress].PlanA[aPath] = aDiff
				out[aAddress].PlanB[aPath] = &Diff{
					Before: "(empty)",
					After:  "(empty)",
				}
				continue
			}

			if *aDiff != *b[aAddress][aPath] {
				// a address and path present in b but diff is not the same

				if _, exists := out[aAddress]; !exists {
					out[aAddress] = CompareEntityDiff{
						PlanA: EntityDiff{},
						PlanB: EntityDiff{},
					}
				}

				out[aAddress].PlanA[aPath] = aDiff
				out[aAddress].PlanB[aPath] = b[aAddress][aPath]
			}
		}
	}

	// Go over each item (B)
	for bAddress, bEntityDiff := range b {

		if _, ok := a[bAddress]; !ok {
			//b address not in a

			out[bAddress] = CompareEntityDiff{
				PlanA: EntityDiff{},
				PlanB: bEntityDiff,
			}

			for path := range bEntityDiff {
				out[bAddress].PlanA[path] = &Diff{
					Before: "(empty)",
					After:  "(empty)",
				}
			}
			continue
		}

		// b address is also in a. Go over each path
		for bPath, bDiff := range bEntityDiff {

			if _, ok := a[bAddress][bPath]; !ok {
				// b path for address is not in a address

				if _, exists := out[bAddress]; !exists {
					out[bAddress] = CompareEntityDiff{
						PlanA: EntityDiff{},
						PlanB: EntityDiff{},
					}
				}

				out[bAddress].PlanA[bPath] = &Diff{
					Before: "(empty)",
					After:  "(empty)",
				}
				out[bAddress].PlanB[bPath] = bDiff
				continue
			}

			if *bDiff != *a[bAddress][bPath] {
				// b address and path present in a but diff is not the same

				if _, exists := out[bAddress]; !exists {
					out[bAddress] = CompareEntityDiff{
						PlanA: EntityDiff{},
						PlanB: EntityDiff{},
					}
				}

				out[bAddress].PlanA[bPath] = a[bAddress][bPath]
				out[bAddress].PlanB[bPath] = bDiff
			}
		}

	}

	return out
}

/*
Compares the output of two Inspects against each other to produce
a comparison output. The output contains diffs only when there is
a divergence between InspectOutput a and b. In other words, only
actual differences are found.
*/
func CompareInspects(a, b *InspectOutput) *CompareInspectsOutput {
	out := &CompareInspectsOutput{
		Diff: &CompareDiff{
			Resources:      getEntityDiffDivergence(a.Diff.Resources, b.Diff.Resources),
			Outputs:        getEntityDiffDivergence(a.Diff.Outputs, b.Diff.Outputs),
			ResourceDrifts: getEntityDiffDivergence(a.Diff.ResourceDrifts, b.Diff.ResourceDrifts),
		},
	}

	return out
}

/*
Produces a slice of strings output which can be printed line by line
to get a Terraform-style stdout report of the compare.
*/
func (c *CompareInspectsOutput) Pretty() []string {
	var out []string
	out = append(out, "\tTerraform plans differ at the following un-filtered changes:\n")

	for address, compEntityDiff := range c.Diff.Resources {
		out = append(out, fmt.Sprintf("\n\t\tresource %s\"%s\"%s changes:\n", colorBold, address, colorNone))
		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan A:%s\n", colorBold, colorNone))

		maxWidth := 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanA {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}

		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan B:%s\n", colorBold, colorNone))

		maxWidth = 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanB {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	for address, compEntityDiff := range c.Diff.ResourceDrifts {
		out = append(out, fmt.Sprintf("\n\t\tresource %s\"%s\"%s drift:\n", colorBold, address, colorNone))
		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan A:%s\n", colorBold, colorNone))

		maxWidth := 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanA {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}

		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan B:%s\n", colorBold, colorNone))

		maxWidth = 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanB {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	for address, compEntityDiff := range c.Diff.Outputs {
		out = append(out, fmt.Sprintf("\n\t\toutput %s\"%s\"%s changes:\n", colorBold, address, colorNone))
		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan A:%s\n", colorBold, colorNone))

		maxWidth := 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanA {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}

		out = append(out, fmt.Sprintf("\n\t\t\t%sPlan B:%s\n", colorBold, colorNone))

		maxWidth = 0
		for path := range compEntityDiff.PlanA {
			if len(path) > maxWidth {
				maxWidth = len(path)
			}
		}

		for path, diff := range compEntityDiff.PlanB {
			out = append(out, fmt.Sprintf("\t\t\t\t%s:%s%s %s->%s %s\n", path, helpers.FillWithSpaces(path, maxWidth), diff.Before, colorOrange, colorNone, diff.After))
		}
	}

	out = append(out, fmt.Sprintf("\n\tChanges: %v resources, %v resource drifts, %v outputs\n", len(c.Diff.Resources), len(c.Diff.ResourceDrifts), len(c.Diff.Outputs)))
	return out
}
