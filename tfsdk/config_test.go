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

func TestConfigGet(t *testing.T) {
	t.Parallel()

	type testConfigGetData struct {
		Name types.String `tfsdk:"name"`
	}

	type testCase struct {
		config        Config
		expected      testConfigGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			config: Config{
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
			expected: testConfigGetData{
				Name: types.String{Value: "namevalue"},
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testConfigGetData

			diags := tc.config.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGet_testTypes(t *testing.T) {
	t.Parallel()

	type testConfigGetData struct {
		Name testtypes.String `tfsdk:"name"`
	}

	type testCase struct {
		config        Config
		expected      testConfigGetData
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"AttrTypeWithValidateError": {
			config: Config{
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
			expected: testConfigGetData{
				Name: testtypes.String{String: types.String{Value: ""}, CreatedBy: testtypes.StringTypeWithValidateError{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestErrorDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
		"AttrTypeWithValidateWarning": {
			config: Config{
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
			expected: testConfigGetData{
				Name: testtypes.String{String: types.String{Value: "namevalue"}, CreatedBy: testtypes.StringTypeWithValidateWarning{}},
			},
			expectedDiags: diag.Diagnostics{testtypes.TestWarningDiagnostic(tftypes.NewAttributePath().WithAttributeName("name"))},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var val testConfigGetData

			diags := tc.config.Get(context.Background(), &val)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGetAttribute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config        Config
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"string": {
			config: Config{
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
			config: Config{
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
			config: Config{
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
			config: Config{
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
			config: Config{
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
			config: Config{
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
			config: Config{
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

			diags := tc.config.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"), tc.target)

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(tc.target, tc.expected, cmp.Transformer("testtypes", func(in *testtypes.String) testtypes.String { return *in }), cmp.Transformer("types", func(in *types.String) types.String { return *in })); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestConfigGetAttributeValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		config        Config
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	testCases := map[string]testCase{
		"basic": {
			config: Config{
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
			config: Config{
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
			config: Config{
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

			val, diags := tc.config.getAttributeValue(context.Background(), tftypes.NewAttributePath().WithAttributeName("name"))

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(val, tc.expected); diff != "" {
				t.Errorf("unexpected value (+wanted, -got): %s", diff)
			}
		})
	}
}
