package plan

import (
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	tfJson "github.com/hashicorp/terraform-json"
	"github.com/orange-car/tfplan/internal/testing/diff"
	"github.com/stretchr/testify/assert"
)

func Test_InspectWithoutWildcard(t *testing.T) {
	cases := map[string]struct {
		plan           *Plan
		input          *InspectInput
		expectedOutput *InspectOutput
		expectedError  error
	}{
		"1 resource filtered out": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address: "aws_instance.example",
						Change: &tfJson.Change{
							Before: map[string]any{
								"ami":           "ami-0397850",
								"instance_type": "t2.medium",
							},
							After: map[string]any{
								"ami":           "ami-12345678",
								"instance_type": "t2.micro",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					ResourceChanges: []Filter{
						{
							NamePattern: "aws_instance.example",
							DiffPatterns: map[string][]Diff{
								".ami": {
									{Before: "ami-0397850", After: "ami-12345678"},
								},
								".instance_type": {
									{Before: "t2.medium", After: "t2.micro"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources:      map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"1 drift filtered out": {
			plan: &Plan{
				ResourceDrift: []*tfJson.ResourceChange{
					{
						Address: "aws_instance.example",
						Change: &tfJson.Change{
							Before: map[string]any{
								"ami":           "ami-0397850",
								"instance_type": "t2.medium",
							},
							After: map[string]any{
								"ami":           "ami-12345678",
								"instance_type": "t2.micro",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					DriftChanges: []Filter{
						{
							NamePattern: "aws_instance.example",
							DiffPatterns: map[string][]Diff{
								".ami": {
									{Before: "ami-0397850", After: "ami-12345678"},
								},
								".instance_type": {
									{Before: "t2.medium", After: "t2.micro"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources:      map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"1 output filtered out": {
			plan: &Plan{
				OutputChanges: map[string]*tfJson.Change{
					"foo-output": {
						Before: "this",
						After:  "that",
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					OutputChanges: []Filter{
						{
							NamePattern: "foo-output",
							DiffPatterns: map[string][]Diff{
								".": {
									{Before: "this", After: "that"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources:      map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"1 resource no filtered": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address: "aws_s3_bucket.this",
						Change: &tfJson.Change{
							Before: map[string]any{
								"bucket": "foo",
							},
							After: map[string]any{
								"bucket": "bar",
								"baz":    "box",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					ResourceChanges: []Filter{
						{
							NamePattern: "aws_cloudwatch_log_group.this",
							DiffPatterns: map[string][]Diff{
								".retention": {
									{Before: "7", After: "10"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
							".baz":    {Before: "(empty)", After: "box"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"1 drift no filtered": {
			plan: &Plan{
				ResourceDrift: []*tfJson.ResourceChange{
					{
						Address: "aws_s3_bucket.this",
						Change: &tfJson.Change{
							Before: map[string]any{
								"bucket": "foo",
								"baz":    "bin",
							},
							After: map[string]any{
								"bucket": "bar",
								"baz":    "box",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					DriftChanges: []Filter{
						{
							NamePattern: "aws_cloudwatch_log_group.this",
							DiffPatterns: map[string][]Diff{
								".retention": {
									{Before: "7", After: "10"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
							".baz":    {Before: "bin", After: "box"},
						},
					},
					Resources: map[string]EntityDiff{},
					Outputs:   map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"1 output no filtered": {
			plan: &Plan{
				OutputChanges: map[string]*tfJson.Change{
					"foo-output": {
						Before: "this",
						After:  "that",
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					OutputChanges: []Filter{
						{
							NamePattern: "bar-output",
							DiffPatterns: map[string][]Diff{
								".": {
									{Before: "1", After: "2"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					Resources:      map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"2 resources 1 filtered": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address: "aws_s3_bucket.this",
						Change: &tfJson.Change{
							Before: map[string]any{
								"bucket": "foo",
							},
							After: map[string]any{
								"bucket": "bar",
							},
						},
					},
					{
						Address: "aws_s3_bucket.that",

						Change: &tfJson.Change{
							Before: map[string]any{
								"bucket": "foo",
							},
							After: map[string]any{
								"bucket": "bar",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					ResourceChanges: []Filter{
						{
							NamePattern: "aws_s3_bucket.this",
							DiffPatterns: map[string][]Diff{
								".bucket": {
									{Before: "foo", After: "bar"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_s3_bucket.that": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"no filter with before sensitives": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address:      "aws_cloudwatch_log_group.this",
						Mode:         "managed",
						Type:         "aws_cloudwatch_log_group",
						Name:         "this",
						ProviderName: "registry.terraform.io/hashicorp/aws",
						Change: &tfJson.Change{
							Before: map[string]any{
								"name": "super-sensitive-string",
							},
							After: map[string]any{
								"name": "foo-skfghsjfhgsjfh",
							},
							AfterUnknown: false,
							BeforeSensitive: map[string]any{
								"name": true,
							},
							AfterSensitive: false,
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_cloudwatch_log_group.this": {
							".name": {Before: "(sensitive value)", After: "foo-skfghsjfhgsjfh"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"no filter with after unknowns": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address:      "aws_cloudwatch_log_group.this",
						Mode:         "managed",
						Type:         "aws_cloudwatch_log_group",
						Name:         "this",
						ProviderName: "registry.terraform.io/hashicorp/aws",
						Change: &tfJson.Change{
							After: map[string]any{
								"kms_key_id":        nil,
								"name":              "foo-skfghsjfhgsjfh",
								"retention_in_days": float64(0),
								"skip_destroy":      false,
								"tags":              nil,
							},
							AfterUnknown: map[string]any{
								"arn":             true,
								"id":              true,
								"log_group_class": true,
								"name_prefix":     true,
								"tags_all":        true,
							},
							BeforeSensitive: false,
							AfterSensitive:  false,
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_cloudwatch_log_group.this": {
							".name":              {Before: "(empty)", After: "foo-skfghsjfhgsjfh"},
							".name_prefix":       {Before: "(empty)", After: "(known after apply)"},
							".log_group_class":   {Before: "(empty)", After: "(known after apply)"},
							".arn":               {Before: "(empty)", After: "(known after apply)"},
							".id":                {Before: "(empty)", After: "(known after apply)"},
							".retention_in_days": {Before: "(empty)", After: "0"},
							".skip_destroy":      {Before: "(empty)", After: "false"},
							".tags_all":          {Before: "(empty)", After: "(known after apply)"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"no filter all known no sensitives": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address:      "aws_cloudwatch_log_group.this",
						Mode:         "managed",
						Type:         "aws_cloudwatch_log_group",
						Name:         "this",
						ProviderName: "registry.terraform.io/hashicorp/aws",
						Change: &tfJson.Change{
							After: map[string]any{
								"name": "foo-skfghsjfhgsjfh",
							},
							AfterUnknown: map[string]any{
								"name": false,
							},
							BeforeSensitive: false,
							AfterSensitive: map[string]any{
								"name": false,
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_cloudwatch_log_group.this": {
							".name": {Before: "(empty)", After: "foo-skfghsjfhgsjfh"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut, gotError := tst.plan.Inspect(tst.input)

			assert.Equal(t, tst.expectedError, gotError)
			diff.Check(t, tst.expectedOutput, gotOut, cmpopts.IgnoreUnexported(Plan{}))
		})
	}
}

func Test_InspectWithWildcard(t *testing.T) {
	cases := map[string]struct {
		plan           *Plan
		input          *InspectInput
		expectedOutput *InspectOutput
		expectedError  error
	}{
		"filter match star wildcard": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address: "aws_instance.example",
						Change: &tfJson.Change{
							Before: map[string]any{
								"ami":           "ami-0397850",
								"instance_type": "t2.medium",
							},
							After: map[string]any{
								"ami":           "ami-12345678",
								"instance_type": "t2.micro",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					ResourceChanges: []Filter{
						{
							NamePattern: "aws_instance.*",
							DiffPatterns: map[string][]Diff{
								".ami": {
									{Before: "ami-0397850", After: "ami-*"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.example": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
		"filter match questionmark wildcard": {
			plan: &Plan{
				ResourceChanges: []*tfJson.ResourceChange{
					{
						Address: "aws_instance.example",
						Change: &tfJson.Change{
							Before: map[string]any{
								"ami":           "ami-0397850",
								"instance_type": "t2.medium",
							},
							After: map[string]any{
								"ami":           "ami-12345678",
								"instance_type": "t2.micro",
							},
						},
					},
				},
			},
			input: &InspectInput{
				Filter: &InspectFilter{
					ResourceChanges: []Filter{
						{
							NamePattern: "aws_instance.?xample",
							DiffPatterns: map[string][]Diff{
								".ami": {
									{Before: "ami-0397850", After: "ami-*"},
								},
							},
						},
					},
				},
			},
			expectedOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.example": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedError: nil,
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut, gotError := tst.plan.Inspect(tst.input)

			assert.Equal(t, tst.expectedError, gotError)
			diff.Check(t, tst.expectedOutput, gotOut, cmpopts.IgnoreUnexported(Plan{}))
		})
	}
}

func Test_InspectPretty(t *testing.T) {
	cases := map[string]struct {
		inspectOutput  *InspectOutput
		expectedOutput []string
	}{
		"simple": {
			inspectOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.example": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedOutput: []string{
				"\tTerraform plan contained the following un-filtered changes:\n",
				"\n\t\tresource \x1b[1m\"aws_instance.example\"\x1b[0m changes:\n",
				"\t\t\t.instance_type: t2.medium \x1b[33m->\x1b[0m t2.micro\n",
				"\n\tChanges: 1 resources, 0 resource drifts, 0 outputs\n",
			},
		},
		"simple with empty before": {
			inspectOutput: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.example": {
							".instance_type": {Before: "(empty)", After: "t2.micro"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{},
					Outputs:        map[string]EntityDiff{},
				},
			},
			expectedOutput: []string{
				"\tTerraform plan contained the following un-filtered changes:\n",
				"\n\t\tresource \x1b[1m\"aws_instance.example\"\x1b[0m changes:\n",
				"\t\t\t.instance_type: (empty) \x1b[33m->\x1b[0m t2.micro\n",
				"\n\tChanges: 1 resources, 0 resource drifts, 0 outputs\n",
			},
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut := tst.inspectOutput.Pretty()
			diff.Check(t, tst.expectedOutput, gotOut)
		})
	}
}
