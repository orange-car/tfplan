package plan

import (
	"testing"

	"github.com/orange-car/tfplan/internal/testing/diff"
)

func Test_CompareInspects(t *testing.T) {
	cases := map[string]struct {
		a              *InspectOutput
		b              *InspectOutput
		expectedOutput *CompareInspectsOutput
	}{
		"no differences": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources:      map[string]CompareEntityDiff{},
					Outputs:        map[string]CompareEntityDiff{},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
		},
		"drift differences": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket":   {Before: "foo", After: "bar"},
							".my_field": {Before: "1", After: "2"},
						},
					},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{},
					Outputs:   map[string]CompareEntityDiff{},
					ResourceDrifts: map[string]CompareEntityDiff{
						"aws_s3_bucket.this": {
							PlanA: EntityDiff{
								".my_field": {Before: "1", After: "2"},
							},
							PlanB: EntityDiff{
								".my_field": {Before: "(empty)", After: "(empty)"},
							},
						},
					},
				},
			},
		},
		"output differences": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
						"this-is-a-missing-output": {
							".": {Before: "biz", After: "box"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{},
					Outputs: map[string]CompareEntityDiff{
						"this-is-a-missing-output": {
							PlanA: EntityDiff{
								".": {Before: "(empty)", After: "(empty)"},
							},
							PlanB: EntityDiff{
								".": {Before: "biz", After: "box"},
							},
						},
					},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
		},
		"resource differences - extra field change": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
							".some_path":     {Before: "was", After: "now"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{
						"aws_instance.this": {
							PlanA: EntityDiff{
								".some_path": &Diff{Before: "was", After: "now"},
							},
							PlanB: EntityDiff{
								".some_path": {Before: "(empty)", After: "(empty)"},
							},
						},
					},
					Outputs:        map[string]CompareEntityDiff{},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
		},
		"resource differences - extra resource": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
						"aws_instance.that": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs: map[string]EntityDiff{
						"foo-output": {
							".": {Before: "this", After: "that"},
						},
					},
					ResourceDrifts: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "foo", After: "bar"},
						},
					},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{
						"aws_instance.that": {
							PlanA: EntityDiff{
								".instance_type": {Before: "t2.medium", After: "t2.micro"},
							},
							PlanB: EntityDiff{
								".instance_type": {Before: "(empty)", After: "(empty)"},
							},
						},
					},
					Outputs:        map[string]CompareEntityDiff{},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
		},
		"resource differences - no overlap": {
			a: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_instance.this": {
							".instance_type": {Before: "t2.medium", After: "t2.micro"},
						},
					},
					Outputs:        map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
				},
			},
			b: &InspectOutput{
				Diff: &InspectDiff{
					Resources: map[string]EntityDiff{
						"aws_s3_bucket.this": {
							".bucket": {Before: "(empty)", After: "orange-car"},
						},
					},
					Outputs:        map[string]EntityDiff{},
					ResourceDrifts: map[string]EntityDiff{},
				},
			},
			expectedOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{
						"aws_instance.this": {
							PlanA: EntityDiff{".instance_type": {Before: "t2.medium", After: "t2.micro"}},
							PlanB: EntityDiff{".instance_type": {Before: "(empty)", After: "(empty)"}},
						},
						"aws_s3_bucket.this": {
							PlanA: EntityDiff{".bucket": {Before: "(empty)", After: "(empty)"}},
							PlanB: EntityDiff{".bucket": {Before: "(empty)", After: "orange-car"}},
						},
					},
					Outputs:        map[string]CompareEntityDiff{},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut := CompareInspects(tst.a, tst.b)
			diff.Check(t, tst.expectedOutput, gotOut)
		})
	}
}

func Test_ComparePretty(t *testing.T) {
	cases := map[string]struct {
		compareInspectsOutput *CompareInspectsOutput
		expectedOutput        []string
	}{
		// "simple path difference": {
		// 	compareInspectsOutput: &CompareInspectsOutput{
		// 		Diff: &CompareDiff{
		// 			Resources: map[string]CompareEntityDiff{
		// 				"resource.this": {
		// 					PlanA: EntityDiff{
		// 						".path": {Before: "foo", After: "bar"},
		// 					},
		// 					PlanB: EntityDiff{
		// 						".path": {Before: "baz", After: "box"},
		// 					},
		// 				},
		// 			},
		// 			Outputs:        map[string]CompareEntityDiff{},
		// 			ResourceDrifts: map[string]CompareEntityDiff{},
		// 		},
		// 	},
		// 	expectedOutput: []string{
		// 		"\tTerraform plans differ at the following un-filtered changes:\n",
		// 		"\n\t\tresource \x1b[1m\"resource.this\"\x1b[0m changes:\n",
		// 		"\n\t\t\t\x1b[1mPlan A:\x1b[0m\n",
		// 		"\t\t\t\t.path: foo \x1b[33m->\x1b[0m bar\n",
		// 		"\n\t\t\t\x1b[1mPlan B:\x1b[0m\n",
		// 		"\t\t\t\t.path: baz \x1b[33m->\x1b[0m box\n",
		// 		"\n\tChanges: 1 resources, 0 resource drifts, 0 outputs\n",
		// 	},
		// },
		// "path absent": {
		// 	compareInspectsOutput: &CompareInspectsOutput{
		// 		Diff: &CompareDiff{
		// 			Resources: map[string]CompareEntityDiff{
		// 				"resource.this": {
		// 					PlanA: EntityDiff{
		// 						".path": {Before: "foo", After: "bar"},
		// 					},
		// 					PlanB: EntityDiff{
		// 						".path": {Before: "(empty)", After: "(empty)"},
		// 					},
		// 				},
		// 			},
		// 			Outputs:        map[string]CompareEntityDiff{},
		// 			ResourceDrifts: map[string]CompareEntityDiff{},
		// 		},
		// 	},
		// 	expectedOutput: []string{
		// 		"\tTerraform plans differ at the following un-filtered changes:\n",
		// 		"\n\t\tresource \x1b[1m\"resource.this\"\x1b[0m changes:\n",
		// 		"\n\t\t\t\x1b[1mPlan A:\x1b[0m\n",
		// 		"\t\t\t\t.path: foo \x1b[33m->\x1b[0m bar\n",
		// 		"\n\t\t\t\x1b[1mPlan B:\x1b[0m\n",
		// 		"\t\t\t\t.path: (empty) \x1b[33m->\x1b[0m (empty)\n",
		// 		"\n\tChanges: 1 resources, 0 resource drifts, 0 outputs\n",
		// 	},
		// },
		// "different resources": {
		// 	compareInspectsOutput: &CompareInspectsOutput{
		// 		Diff: &CompareDiff{
		// 			Resources: map[string]CompareEntityDiff{
		// 				"resource.this": {
		// 					PlanA: EntityDiff{
		// 						".path": {Before: "foo", After: "bar"},
		// 					},
		// 					PlanB: EntityDiff{},
		// 				},
		// 			},
		// 			Outputs:        map[string]CompareEntityDiff{},
		// 			ResourceDrifts: map[string]CompareEntityDiff{},
		// 		},
		// 	},
		// 	expectedOutput: []string{
		// 		"\tTerraform plans differ at the following un-filtered changes:\n",
		// 		"\n\t\tresource \x1b[1m\"resource.this\"\x1b[0m changes:\n",
		// 		"\n\t\t\t\x1b[1mPlan A:\x1b[0m\n",
		// 		"\t\t\t\t.path: foo \x1b[33m->\x1b[0m bar\n",
		// 		"\n\t\t\t\x1b[1mPlan B:\x1b[0m\n",
		// 		"\n\tChanges: 1 resources, 0 resource drifts, 0 outputs\n",
		// 	},
		// },
		"missing output": {
			compareInspectsOutput: &CompareInspectsOutput{
				Diff: &CompareDiff{
					Resources: map[string]CompareEntityDiff{},
					Outputs: map[string]CompareEntityDiff{
						"my-output": {
							PlanA: EntityDiff{
								".": {Before: "(empty)", After: "(empty)"},
							},
							PlanB: EntityDiff{
								".": {Before: "0", After: "1"},
							},
						},
					},
					ResourceDrifts: map[string]CompareEntityDiff{},
				},
			},
			expectedOutput: []string{
				"\tTerraform plans differ at the following un-filtered changes:\n",
				"\n\t\toutput \x1b[1m\"my-output\"\x1b[0m changes:\n",
				"\n\t\t\t\x1b[1mPlan A:\x1b[0m\n",
				"\t\t\t\t.: (empty) \x1b[33m->\x1b[0m (empty)\n",
				"\n\t\t\t\x1b[1mPlan B:\x1b[0m\n",
				"\t\t\t\t.: 0 \x1b[33m->\x1b[0m 1\n",
				"\n\tChanges: 0 resources, 0 resource drifts, 1 outputs\n",
			},
		},
	}

	for name, tst := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			gotOut := tst.compareInspectsOutput.Pretty()
			diff.Check(t, tst.expectedOutput, gotOut)
		})
	}
}
