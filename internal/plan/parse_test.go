package plan

import (
	"fmt"
	"testing"

	tfJson "github.com/hashicorp/terraform-json"
	"github.com/orange-car/tfplan/internal/testing/diff"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
)

func Test_ParsePlan(t *testing.T) {

	cases := map[string]struct {
		jsonPlan       []byte
		expectedOutput *Plan
		expectedError  error
	}{
		"no error": {
			jsonPlan: []byte(`
			{
				"resource_changes": [
					{
						"address": "aws_cloudwatch_log_group.this",
						"mode": "managed",
						"type": "aws_cloudwatch_log_group",
						"name": "this",
						"provider_name": "registry.terraform.io/hashicorp/aws",
						"change": {
							"before": null,
							"after": {
								"kms_key_id": null,
								"name": "foo-skfghsjfhgsjfh",
								"retention_in_days": 0,
								"skip_destroy": false,
								"tags": null
							},
							"after_unknown": {
								"arn": true,
								"id": true,
								"log_group_class": true,
								"name_prefix": true,
								"tags_all": true
							},
							"before_sensitive": false,
							"after_sensitive": {
								"tags_all": {}
							}
						}
					}
				]
			}`),
			expectedOutput: &Plan{
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
							AfterSensitive:  map[string]any{"tags_all": map[string]any{}},
						},
					},
				},
			},
			expectedError: nil,
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut, gotError := ParsePlan(tst.jsonPlan)

			assert.Equal(t, tst.expectedError, gotError)
			diff.Check(t, tst.expectedOutput, gotOut, cmpopts.IgnoreUnexported(Plan{}))
		})
	}
}

func Test_ParseInspectFilter(t *testing.T) {

	cases := map[string]struct {
		jsonFilter     []byte
		expectedOutput *InspectFilter
		expectedError  error
	}{
		"no error": {
			jsonFilter: []byte(`
				{
					"outputChanges": [
						{
							"namePattern": "foo_output_name",
							"diffPatterns": {
								"._foo_output_path": [
									{
										"before": "foo_output_before",
										"after": "foo_output_after"
									}
								]
							}
						}
					],
					"resourceChanges": [
						{
							"namePattern": "foo_resource_name",
							"diffPatterns": {
								"._foo_resource_path": [
									{
										"before": "foo_resource_before",
										"after": "foo_resource_after"
									}
								]
							}
						}
					],
					"driftChanges": [
						{
							"namePattern": "foo_drift_resource_name",
							"diffPatterns": {
								"._foo_drift_resource_path": [
									{
										"before": "foo_drift_resource_before",
										"after": "foo_drift_resource_after"
									}
								]
							}
						}
					]
				}
			`),
			expectedOutput: &InspectFilter{
				OutputChanges: []Filter{
					{
						NamePattern: "foo_output_name",
						DiffPatterns: map[string][]Diff{
							"._foo_output_path": {
								{
									Before: "foo_output_before",
									After:  "foo_output_after",
								},
							},
						},
					},
				},
				ResourceChanges: []Filter{
					{
						NamePattern: "foo_resource_name",
						DiffPatterns: map[string][]Diff{
							"._foo_resource_path": {
								{
									Before: "foo_resource_before",
									After:  "foo_resource_after",
								},
							},
						},
					},
				},
				DriftChanges: []Filter{
					{
						NamePattern: "foo_drift_resource_name",
						DiffPatterns: map[string][]Diff{
							"._foo_drift_resource_path": {
								{
									Before: "foo_drift_resource_before",
									After:  "foo_drift_resource_after",
								},
							},
						},
					},
				},
			},
			expectedError: nil,
		},
		"json error": {
			jsonFilter:     []byte(``),
			expectedOutput: nil,
			expectedError:  fmt.Errorf("unable to unmarshal inspect filter caused by: unexpected end of JSON input"),
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut, gotError := ParseInspectFilter(tst.jsonFilter)

			assert.Equal(t, tst.expectedError, gotError)
			diff.Check(t, tst.expectedOutput, gotOut, cmpopts.IgnoreUnexported(Plan{}))
		})
	}
}
