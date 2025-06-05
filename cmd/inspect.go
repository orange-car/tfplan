package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/orange-car/tfplan/internal/plan"

	"github.com/spf13/cobra"
)

type inspectPlanInput struct {
	tfplan           *plan.Plan
	filter           *plan.InspectFilter
	prettyPrint      bool
	detailedExitCode bool
}

func inspectPlan(in *inspectPlanInput) error {

	if in.tfplan == nil {
		return fmt.Errorf("plan cannot be empty")
	}

	out, err := in.tfplan.Inspect(&plan.InspectInput{
		Filter: in.filter,
	})
	if err != nil {
		return err
	}

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

// inspectCmd represents the inspect command
var inspectCmd = &cobra.Command{
	Use:   "inspect",
	Args:  cobra.MaximumNArgs(1),
	Short: "Check a plan for expected changes",
	Long: `
Inspects a JSON Terraform plan for changes to outputs, resource and resource drift. 
Compares changes against your provided filter criteria to filter out changes you
don't care about. For any changes not captured by your filter, they will be printed.

Filtering Terraform plans is useful when working with Terraform programmatically such as
in a deployment pipeline. For example, you may have a step to inspect the plan and explicitly
exit or add a manual approval step if there are changes you are not allow-listing. A common
example of this may be a change to an aws_lambda_function based on the source_code_hash. In this
case, you may wish to auto-approve the Terraform plan. For any other changes, you may wish to seek manual 
approval before proceeding. 

Example usage:
$ tfplan inspect \
--plan "$(terraform show --json .plan)" \
--detailed-exitcode \
--filter "$(cat filter.json)" \
--pretty
`,
	PreRunE: nil,
	RunE: func(cmd *cobra.Command, args []string) error {

		planFlg, err := cmd.Flags().GetString("plan")
		if err != nil {
			return fmt.Errorf("failed to get plan flag caused by: %v", err)
		}

		tfplan, err := plan.ParsePlan([]byte(planFlg))
		if err != nil {
			return err
		}

		filterFlg, err := cmd.Flags().GetString("filter")
		if err != nil {
			return fmt.Errorf("failed to get filter flag caused by: %v", err)
		}

		filter := &plan.InspectFilter{}
		if filterFlg != "" && filterFlg != "{}" {
			filter, err = plan.ParseInspectFilter([]byte(filterFlg))
			if err != nil {
				return err
			}
		}

		detailedFlg, err := cmd.Flags().GetBool("detailed-exitcode")
		if err != nil {
			return fmt.Errorf("failed to get detailed-exitcode flag caused by: %v", err)
		}

		prettyFlg, err := cmd.Flags().GetBool("pretty")
		if err != nil {
			return fmt.Errorf("failed to get pretty flag caused by: %v", err)
		}

		return inspectPlan(&inspectPlanInput{
			tfplan:           tfplan,
			filter:           filter,
			prettyPrint:      prettyFlg,
			detailedExitCode: detailedFlg,
		})
	},
}

func init() {
	rootCmd.AddCommand(inspectCmd)
	inspectCmd.PersistentFlags().StringP("plan", "p", "", "plan (json format) to inspect")
	inspectCmd.PersistentFlags().StringP("filter", "f", "{}", "filter (json format) to filter out changes")
	inspectCmd.PersistentFlags().BoolP("detailed-exitcode", "d", false, "when used, exit code 2 will return when there are unfiltered changes")
	inspectCmd.PersistentFlags().BoolP("pretty", "P", false, "print the results in a human readable format")

	// Required flags
	inspectCmd.MarkFlagRequired("plan")
}
