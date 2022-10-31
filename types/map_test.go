package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestMapTypeTerraformType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    MapType
		expected tftypes.Type
	}
	tests := map[string]testCase{
		"map-of-strings": {
			input: MapType{
				ElemType: StringType,
			},
			expected: tftypes.Map{
				ElementType: tftypes.String,
			},
		},
		"map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: StringType,
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.String,
				},
			},
		},
		"map-of-map-of-map-of-strings": {
			input: MapType{
				ElemType: MapType{
					ElemType: MapType{
						ElemType: StringType,
					},
				},
			},
			expected: tftypes.Map{
				ElementType: tftypes.Map{
					ElementType: tftypes.Map{
						ElementType: tftypes.String,
					},
				},
			},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			got := test.input.TerraformType(context.Background())
			if !got.Equal(test.expected) {
				t.Errorf("Expected %s, got %s", test.expected, got)
			}
		})
	}
}

func TestMapTypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver    MapType
		input       tftypes.Value
		expected    attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"basic-map": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, map[string]tftypes.Value{
				"one":   tftypes.NewValue(tftypes.Number, 1),
				"two":   tftypes.NewValue(tftypes.Number, 2),
				"three": tftypes.NewValue(tftypes.Number, 3),
			}),
			expected: MapValueMust(
				NumberType,
				map[string]attr.Value{
					"one":   NumberValue(big.NewFloat(1)),
					"two":   NumberValue(big.NewFloat(2)),
					"three": NumberValue(big.NewFloat(3)),
				},
			),
		},
		"wrong-type": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input:       tftypes.NewValue(tftypes.String, "wrong"),
			expectedErr: `can't use tftypes.String<"wrong"> as value of Map, can only use tftypes.Map values`,
		},
		"nil-type": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input:    tftypes.NewValue(nil, nil),
			expected: MapNull(NumberType),
		},
		"unknown": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, tftypes.UnknownValue),
			expected: MapUnknown(NumberType),
		},
		"null": {
			receiver: MapType{
				ElemType: NumberType,
			},
			input: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.Number,
			}, nil),
			expected: MapNull(NumberType),
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.receiver.ValueFromTerraform(context.Background(), test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err.Error())
					return
				}
				if err.Error() != test.expectedErr {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
			}
			if test.expectedErr != "" && err == nil {
				t.Errorf("Expected err to be %q, got nil", test.expectedErr)
				return
			}
			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("unexpected result (-expected, +got): %s", diff)
			}
			if test.expected != nil && test.expected.IsNull() != test.input.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", test.expected.IsNull(), test.input.IsNull())
			}
			if test.expected != nil && test.expected.IsUnknown() != !test.input.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", test.expected.IsUnknown(), !test.input.IsKnown())
			}
		})
	}
}

func TestMapTypeEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver MapType
		input    attr.Type
		expected bool
	}
	tests := map[string]testCase{
		"equal": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			expected: true,
		},
		"diff": {
			receiver: MapType{
				ElemType: ListType{
					ElemType: StringType,
				},
			},
			input: MapType{
				ElemType: ListType{
					ElemType: NumberType,
				},
			},
			expected: false,
		},
		"wrongType": {
			receiver: MapType{
				ElemType: StringType,
			},
			input:    NumberType,
			expected: false,
		},
		"nil": {
			receiver: MapType{
				ElemType: StringType,
			},
			input:    nil,
			expected: false,
		},
		"nil-elem": {
			receiver: MapType{},
			input:    MapType{},
			// MapTypes with nil ElemTypes are invalid, and aren't
			// equal to anything
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if test.expected != got {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestMapValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      map[string]attr.Value
		expected      Map
		expectedDiags diag.Diagnostics
	}{
		"valid-no-elements": {
			elementType: StringType,
			elements:    map[string]attr.Value{},
			expected:    MapValueMust(StringType, map[string]attr.Value{}),
		},
		"valid-elements": {
			elementType: StringType,
			elements: map[string]attr.Value{
				"null":    StringNull(),
				"unknown": StringUnknown(),
				"known":   StringValue("test"),
			},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{
					"null":    StringNull(),
					"unknown": StringUnknown(),
					"known":   StringValue("test"),
				},
			),
		},
		"invalid-element-type": {
			elementType: StringType,
			elements: map[string]attr.Value{
				"string": StringValue("test"),
				"bool":   BoolValue(true),
			},
			expected: MapUnknown(StringType),
			expectedDiags: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Map Element Type",
					"While creating a Map value, an invalid element was detected. "+
						"A Map must use the single, given element type. "+
						"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
						"Map Element Type: types.StringType\n"+
						"Map Key (bool) Element Type: types.BoolType",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := MapValue(testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestMapValueFrom(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		elementType   attr.Type
		elements      any
		expected      Map
		expectedDiags diag.Diagnostics
	}{
		"valid-StringType-map[string]attr.Value-empty": {
			elementType: StringType,
			elements:    map[string]attr.Value{},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{},
			),
		},
		"valid-StringType-map[string]types.String-empty": {
			elementType: StringType,
			elements:    map[string]String{},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{},
			),
		},
		"valid-StringType-map[string]types.String": {
			elementType: StringType,
			elements: map[string]String{
				"key1": StringNull(),
				"key2": StringUnknown(),
				"key3": StringValue("test"),
			},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringNull(),
					"key2": StringUnknown(),
					"key3": StringValue("test"),
				},
			),
		},
		"valid-StringType-map[string]*string": {
			elementType: StringType,
			elements: map[string]*string{
				"key1": nil,
				"key2": pointer("test1"),
				"key3": pointer("test2"),
			},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringNull(),
					"key2": StringValue("test1"),
					"key3": StringValue("test2"),
				},
			),
		},
		"valid-StringType-map[string]string": {
			elementType: StringType,
			elements: map[string]string{
				"key1": "test1",
				"key2": "test2",
			},
			expected: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("test1"),
					"key2": StringValue("test2"),
				},
			),
		},
		"invalid-not-map": {
			elementType: StringType,
			elements:    "oops",
			expected:    MapUnknown(StringType),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Map Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Map value, received tftypes.Value with value: tftypes.String<\"oops\">",
				),
			},
		},
		"invalid-type": {
			elementType: StringType,
			elements:    map[string]bool{"key1": true},
			expected:    MapUnknown(StringType),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty().AtMapKey("key1"),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"can't unmarshal tftypes.Bool into *string, expected string",
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := MapValueFrom(context.Background(), testCase.elementType, testCase.elements)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestMapElementsAs_mapStringString(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]string
	expected := map[string]string{
		"h": "hello",
		"w": "world",
	}

	diags := MapValueMust(
		StringType,
		map[string]attr.Value{
			"h": StringValue("hello"),
			"w": StringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapElementsAs_mapStringAttributeValue(t *testing.T) {
	t.Parallel()

	var stringSlice map[string]String
	expected := map[string]String{
		"h": StringValue("hello"),
		"w": StringValue("world"),
	}

	diags := MapValueMust(
		StringType,
		map[string]attr.Value{
			"h": StringValue("hello"),
			"w": StringValue("world"),
		},
	).ElementsAs(context.Background(), &stringSlice, false)
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if diff := cmp.Diff(stringSlice, expected); diff != "" {
		t.Errorf("Unexpected diff (-expected, +got): %s", diff)
	}
}

func TestMapToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Map
		expectation tftypes.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"known": {
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, "hello"),
				"key2": tftypes.NewValue(tftypes.String, "world"),
			}),
		},
		"known-partial-unknown": {
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringUnknown(),
					"key2": StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				"key2": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"known-partial-null": {
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringNull(),
					"key2": StringValue("hello, world"),
				},
			),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
				"key1": tftypes.NewValue(tftypes.String, nil),
				"key2": tftypes.NewValue(tftypes.String, "hello, world"),
			}),
		},
		"unknown": {
			input:       MapUnknown(StringType),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, tftypes.UnknownValue),
		},
		"null": {
			input:       MapNull(StringType),
			expectation: tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, gotErr := test.input.ToTerraformValue(context.Background())

			if test.expectedErr == "" && gotErr != nil {
				t.Errorf("Unexpected error: %s", gotErr)
				return
			}

			if test.expectedErr != "" {
				if gotErr == nil {
					t.Errorf("Expected error to be %q, got none", test.expectedErr)
					return
				}

				if test.expectedErr != gotErr.Error() {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, gotErr.Error())
					return
				}
			}

			if diff := cmp.Diff(got, test.expectation); diff != "" {
				t.Errorf("Unexpected result (+got, -expected): %s", diff)
			}
		})
	}
}

func TestMapElements(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Map
		expected map[string]attr.Value
	}{
		"known": {
			input:    MapValueMust(StringType, map[string]attr.Value{"test-key": StringValue("test-value")}),
			expected: map[string]attr.Value{"test-key": StringValue("test-value")},
		},
		"null": {
			input:    MapNull(StringType),
			expected: nil,
		},
		"unknown": {
			input:    MapUnknown(StringType),
			expected: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.Elements()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapElementType(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Map
		expected attr.Type
	}{
		"known": {
			input:    MapValueMust(StringType, map[string]attr.Value{"test-key": StringValue("test-value")}),
			expected: StringType,
		},
		"null": {
			input:    MapNull(StringType),
			expected: StringType,
		},
		"unknown": {
			input:    MapUnknown(StringType),
			expected: StringType,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ElementType(context.Background())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		receiver Map
		input    attr.Value
		expected bool
	}
	tests := map[string]testCase{
		"known-known": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			expected: true,
		},
		"known-known-diff-value": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("goodnight"),
					"key2": StringValue("moon"),
				},
			),
			expected: false,
		},
		"known-known-diff-length": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
					"key3": StringValue("extra"),
				},
			),
			expected: false,
		},
		"known-known-diff-type": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input: SetValueMust(
				BoolType,
				[]attr.Value{
					BoolValue(false),
					BoolValue(true),
				},
			),
			expected: false,
		},
		"known-known-diff-unknown": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringUnknown(),
				},
			),
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			expected: false,
		},
		"known-known-diff-null": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringNull(),
				},
			),
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			expected: false,
		},
		"known-unknown": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input:    MapUnknown(StringType),
			expected: false,
		},
		"known-null": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input:    MapNull(StringType),
			expected: false,
		},
		"known-diff-type": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input: SetValueMust(
				StringType,
				[]attr.Value{
					StringValue("hello"),
					StringValue("world"),
				},
			),
			expected: false,
		},
		"known-nil": {
			receiver: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			input:    nil,
			expected: false,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.receiver.Equal(test.input)
			if got != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, got)
			}
		})
	}
}

func TestMapIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Map
		expected bool
	}{
		"known": {
			input:    MapValueMust(StringType, map[string]attr.Value{"test-key": StringValue("test-value")}),
			expected: false,
		},
		"null": {
			input:    MapNull(StringType),
			expected: true,
		},
		"unknown": {
			input:    MapUnknown(StringType),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsNull()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Map
		expected bool
	}{
		"known": {
			input:    MapValueMust(StringType, map[string]attr.Value{"test-key": StringValue("test-value")}),
			expected: false,
		},
		"null": {
			input:    MapNull(StringType),
			expected: false,
		},
		"unknown": {
			input:    MapUnknown(StringType),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsUnknown()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestMapString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Map
		expectation string
	}
	tests := map[string]testCase{
		"known": {
			input: MapValueMust(
				Int64Type,
				map[string]attr.Value{
					"alpha": Int64Value(1234),
					"beta":  Int64Value(56789),
					"gamma": Int64Value(9817),
					"sigma": Int64Value(62534),
				},
			),
			expectation: `{"alpha":1234,"beta":56789,"gamma":9817,"sigma":62534}`,
		},
		"known-map-of-maps": {
			input: MapValueMust(
				MapType{
					ElemType: StringType,
				},
				map[string]attr.Value{
					"first": MapValueMust(
						StringType,
						map[string]attr.Value{
							"alpha": StringValue("hello"),
							"beta":  StringValue("world"),
							"gamma": StringValue("foo"),
							"sigma": StringValue("bar"),
						},
					),
					"second": MapValueMust(
						StringType,
						map[string]attr.Value{
							"echo": StringValue("echo"),
						},
					),
				},
			),
			expectation: `{"first":{"alpha":"hello","beta":"world","gamma":"foo","sigma":"bar"},"second":{"echo":"echo"}}`,
		},
		"known-key-quotes": {
			input: MapValueMust(
				BoolType,
				map[string]attr.Value{
					`testing is "fun"`: BoolValue(true),
				},
			),
			expectation: `{"testing is \"fun\"":true}`,
		},
		"unknown": {
			input:       MapUnknown(StringType),
			expectation: "<unknown>",
		},
		"null": {
			input:       MapNull(StringType),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Map{},
			expectation: "<null>",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.String()
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}

func TestMapType(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Map
		expectation attr.Type
	}
	tests := map[string]testCase{
		"known": {
			input: MapValueMust(
				StringType,
				map[string]attr.Value{
					"key1": StringValue("hello"),
					"key2": StringValue("world"),
				},
			),
			expectation: MapType{ElemType: StringType},
		},
		"known-map-of-maps": {
			input: MapValueMust(
				MapType{
					ElemType: StringType,
				},
				map[string]attr.Value{
					"key1": MapValueMust(
						StringType,
						map[string]attr.Value{
							"key1": StringValue("hello"),
							"key2": StringValue("world"),
						},
					),
					"key2": MapValueMust(
						StringType,
						map[string]attr.Value{
							"key1": StringValue("foo"),
							"key2": StringValue("bar"),
						},
					),
				},
			),
			expectation: MapType{
				ElemType: MapType{
					ElemType: StringType,
				},
			},
		},
		"unknown": {
			input:       MapUnknown(StringType),
			expectation: MapType{ElemType: StringType},
		},
		"null": {
			input:       MapNull(StringType),
			expectation: MapType{ElemType: StringType},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Type(context.Background())
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}

func TestMapTypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		mapType       MapType
		tfValue       tftypes.Value
		path          path.Path
		expectedDiags diag.Diagnostics
	}{
		"wrong-value-type": {
			mapType: MapType{
				ElemType: StringType,
			},
			tfValue: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Map Type Validation Error",
					"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"expected Map value, received tftypes.Value with value: tftypes.List[tftypes.String]<tftypes.String<\"testvalue\">>",
				),
			},
		},
		"no-validation": {
			mapType: MapType{
				ElemType: StringType,
			},
			tfValue: tftypes.NewValue(tftypes.Map{
				ElementType: tftypes.String,
			}, map[string]tftypes.Value{
				"testkey": tftypes.NewValue(tftypes.String, "testvalue"),
			}),
			path: path.Root("test"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.mapType.Validate(context.Background(), testCase.tfValue, testCase.path)

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
