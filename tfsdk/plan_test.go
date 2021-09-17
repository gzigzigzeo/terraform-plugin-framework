package tfsdk

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	intreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPlanGet(t *testing.T) {
	t.Parallel()

	type testPlanGetData struct {
		Name types.String `tfsdk:"name"`
	}

	type testCase struct {
		plan          Plan
		expected      testPlanGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetData{
				Name: types.String{Value: "namevalue"},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testPlanGetData

			diags := tc.plan.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGet_testTypes(t *testing.T) {
	t.Parallel()

	type testPlanGetDataTestTypes struct {
		Name testtypes.String `tfsdk:"name"`
	}

	type testCase struct {
		plan          Plan
		expected      testPlanGetDataTestTypes
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetDataTestTypes{
				Name: testtypes.String{String: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			expected: testPlanGetDataTestTypes{
				Name: testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testPlanGetDataTestTypes

			diags := tc.plan.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"string": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(string),
			expected: newStringPointer("namevalue"),
		},
		"*string": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(*string),
			expected: newStringPointerPointer("namevalue"),
		},
		"types.String": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(types.String),
			expected: &types.String{Value: "namevalue"},
		},
		"incompatible-target": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(testtypes.String),
			expected: new(testtypes.String),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"Value Conversion Error",
					intreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(types.String{Value: "namevalue"}),
						TargetType: reflect.TypeOf(testtypes.String{}),
						AttrPath:   tftypes.NewAttributePath().WithAttributeName("name"),
						SchemaType: types.StringType,
					}.Detail(),
				),
			},
		},
		"incompatible-type": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			target:   new(bool),
			expected: new(bool),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					tftypes.NewAttributePath().WithAttributeName("name"),
					"Value Conversion Error",
					intreflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "namevalue"),
						TargetType: reflect.TypeOf(false),
						Err:        fmt.Errorf("can't unmarshal %s into *%T, expected boolean", tftypes.String, false),
						AttrPath:   tftypes.NewAttributePath().WithAttributeName("name"),
					}.Detail(),
				),
			},
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			target:        new(testtypes.String),
			expected:      new(testtypes.String),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			target:        new(testtypes.String),
			expected:      &testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanGetAttributeValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			expected: types.String{Value: "namevalue"},
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			expected:      nil,
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "namevalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			expected:      testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, diags := tc.plan.getAttributeValue(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSet(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"write": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "oldvalue"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"overwrite": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected:      tftypes.Value{},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.Value{},
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			val: struct {
				Name string `tfsdk:"name"`
			}{
				Name: "newvalue",
			},
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newvalue"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.Set(context.Background(), tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPlanSetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		plan          Plan
		path          *tftypes.AttributePath
		val           interface{}
		expected      tftypes.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     types.StringType,
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
		},
		"AttrTypeWithValidateError": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateError{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "originalname"),
			}),
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			plan: Plan{
				Raw: tftypes.NewValue(tftypes.Object{
					AttributeTypes: map[string]tftypes.Type{
						"name": tftypes.String,
					},
				}, map[string]tftypes.Value{
					"name": tftypes.NewValue(tftypes.String, "originalname"),
				}),
				Schema: Schema{
					Attributes: map[string]Attribute{
						"name": {
							Type:     testtypes.StringTypeWithValidateWarning{},
							Required: true,
						},
					},
				},
			},
			path: tftypes.NewAttributePath().WithAttributeName("name"),
			val:  "newname",
			expected: tftypes.NewValue(tftypes.Object{
				AttributeTypes: map[string]tftypes.Type{
					"name": tftypes.String,
				},
			}, map[string]tftypes.Value{
				"name": tftypes.NewValue(tftypes.String, "newname"),
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name")),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := tc.plan.SetAttribute(context.Background(), tc.path, tc.val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.plan.Raw, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
