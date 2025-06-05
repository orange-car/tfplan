package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orange-car/tfplan/internal/plan"
	"github.com/spf13/cobra"
)

type comparePlanInput struct {
	planA            *plan.Plan
	planB            *plan.Plan
	filter           *plan.InspectFilter
	prettyPrint      bool
	detailedExitCode bool
}

func comparePlans(in *comparePlanInput) error {

	if in.planA == nil {
		return fmt.Errorf("plan-a cannot be empty")
	}

	if in.planB == nil {
		return fmt.Errorf("plan-b cannot be empty")
	}

	aOut, err := in.planA.Inspect(&plan.InspectInput{
		Filter: in.filter,
	})
	if err != nil {
		return err
	}

	bOut, err := in.planB.Inspect(&plan.InspectInput{
		Filter: in.filter,
	})
	if err != nil {
		return err
	}

	out := plan.CompareInspects(aOut, bOut)

	if in.prettyPrint {
		for _, line := range out.Pretty() {
			fmt.Print(line)
		}
	} else {
		bytes, err := json.Marshal(out)
		if err != nil {
			return fmt.Errorf("failed to json marshal inspection output caused by: %v", err)
		}
		fmt.Println(string(bytes))
	}

	if !out.IsEmpty() && in.detailedExitCode {
		os.Exit(2)
	}

	return nil
}

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare two plans for differences",
	Long: `
Inspects two JSON Terraform plans for changes to outputs, resource and resource drift with changes
filtered out by your provided filter criteria. Compares changes against each other and reports 
differences between the two plans.

Comparing plans programmatically is particularly useful when they are large and/or you have to do
it often. Applying the optional filter can be useful to rule out changes you don't care about.

Example usage:
$ tfplan compare \
--plan-a "$(terraform show --json a.plan)" \
--plan-b "$(terraform show --json b.plan)" \
--detailed-exitcode \
--filter "$(cat filter.json)" \
--pretty
`,
	PreRunE: nil,
	// RunE:    compareRunner,
	RunE: func(cmd *cobra.Command, args []string) error {

		planAFlg, err := cmd.Flags().GetString("plan-a")
		if err != nil {
			return fmt.Errorf("failed to get plan-a flag caused by: %v", err)
		}

		planBFlg, err := cmd.Flags().GetString("plan-b")
		if err != nil {
			return fmt.Errorf("failed to get plan-b flag caused by: %v", err)
		}

		tfplanA, err := plan.ParsePlan([]byte(planAFlg))
		if err != nil {
			return err
		}

		tfplanB, err := plan.ParsePlan([]byte(planBFlg))
		if err != nil {
			return err
		}

		filterFlg, err := cmd.Flags().GetString("filter")
		if err != nil {
			return fmt.Errorf("failed to get filter flag caused by: %v", err)
		}

		filter, err := plan.ParseInspectFilter([]byte(filterFlg))
		if err != nil {
			return err
		}

		detailedFlg, err := cmd.Flags().GetBool("detailed-exitcode")
		if err != nil {
			return fmt.Errorf("failed to get detailed-exitcode flag caused by: %v", err)
		}

		prettyFlg, err := cmd.Flags().GetBool("pretty")
		if err != nil {
			return fmt.Errorf("failed to get pretty flag caused by: %v", err)
		}

		return comparePlans(&comparePlanInput{
			planA:            tfplanA,
			planB:            tfplanB,
			filter:           filter,
			prettyPrint:      prettyFlg,
			detailedExitCode: detailedFlg,
		})
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	compareCmd.PersistentFlags().StringP("plan-a", "a", "", "plan (json format) to compare against --plan-b (-b)")
	compareCmd.PersistentFlags().StringP("plan-b", "b", "", "plan (json format) to compare against --plan-a (-a)")
	compareCmd.PersistentFlags().StringP("filter", "f", "{}", "filter (json format) to filter out changes")
	compareCmd.PersistentFlags().BoolP("detailed-exitcode", "d", false, "when used, exit code 2 will return when there are unfiltered changes")
	compareCmd.PersistentFlags().BoolP("pretty", "P", false, "print the results in a human readable format")

	// Required flags
	compareCmd.MarkPersistentFlagRequired("plan-a")
	compareCmd.MarkPersistentFlagRequired("plan-b")
}
