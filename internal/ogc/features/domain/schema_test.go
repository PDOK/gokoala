package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSchema(t *testing.T) {
	tests := []struct {
		name           string
		fields         []Field
		fidColumn      string
		externalFid    string
		expectedError  bool
		expectedErrMsg string
		expectedSchema *Schema
	}{
		{
			name: "valid single geometry field",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "location", Type: "Point"},
				{Name: "name", Type: "string"},
			},
			fidColumn:     "id",
			externalFid:   "",
			expectedError: false,
			expectedSchema: &Schema{
				Fields: []Field{
					{Name: "id", Type: "integer", IsFid: true, IsExternalFid: false},
					{Name: "location", Type: "Point", IsFid: false, IsExternalFid: false},
					{Name: "name", Type: "string", IsFid: false, IsExternalFid: false},
				},
			},
		},
		{
			name: "fail on multiple geometry fields",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "location", Type: "Point"},
				{Name: "shape", Type: "Polygon"},
			},
			fidColumn:      "id",
			externalFid:    "",
			expectedError:  true,
			expectedErrMsg: "more than one geometry field found",
		},
		{
			name: "fail on empty field name",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "", Type: "Point"},
				{Name: "shape", Type: "Polygon"},
			},
			fidColumn:      "id",
			externalFid:    "",
			expectedError:  true,
			expectedErrMsg: "empty field name found, field name is required",
		},
		{
			name: "fail on empty field type",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "location", Type: "Point"},
				{Name: "shape", Type: ""},
			},
			fidColumn:      "id",
			externalFid:    "",
			expectedError:  true,
			expectedErrMsg: "empty field type found, field type is required",
		},
		{
			name: "fail on non-existing external fid",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "location", Type: "Point"},
			},
			fidColumn:      "id",
			externalFid:    "ext_fid", // not present
			expectedError:  true,
			expectedErrMsg: "external feature ID column 'ext_fid' configured but not found in schema",
		},
		{
			name: "fields to skip are ignored",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "minx", Type: "float"},
				{Name: "location", Type: "Point"},
			},
			fidColumn:     "id",
			externalFid:   "",
			expectedError: false,
			expectedSchema: &Schema{
				Fields: []Field{
					{Name: "id", Type: "integer", IsFid: true, IsExternalFid: false},
					{Name: "location", Type: "Point", IsFid: false, IsExternalFid: false},
				},
			},
		},
		{
			name: "valid external FID column",
			fields: []Field{
				{Name: "id", Type: "integer"},
				{Name: "ext_id", Type: "string"},
				{Name: "location", Type: "Point"},
			},
			fidColumn:     "id",
			externalFid:   "ext_id",
			expectedError: false,
			expectedSchema: &Schema{
				Fields: []Field{
					{Name: "id", Type: "integer", IsFid: true, IsExternalFid: false},
					{Name: "ext_id", Type: "string", IsFid: false, IsExternalFid: true},
					{Name: "location", Type: "Point", IsFid: false, IsExternalFid: false},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema, err := NewSchema(tt.fields, tt.fidColumn, tt.externalFid)

			if tt.expectedError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErrMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedSchema, schema)
			}
		})
	}
}

func TestHasExternalFid(t *testing.T) {
	tests := []struct {
		name          string
		schema        Schema
		expectedValue bool
	}{
		{
			name: "no external FID",
			schema: Schema{
				Fields: []Field{
					{Name: "id", Type: "integer", IsExternalFid: false},
					{Name: "name", Type: "string", IsExternalFid: false},
				},
			},
			expectedValue: false,
		},
		{
			name: "has external FID",
			schema: Schema{
				Fields: []Field{
					{Name: "id", Type: "integer", IsExternalFid: false},
					{Name: "ext_id", Type: "string", IsExternalFid: true},
					{Name: "location", Type: "Point", IsExternalFid: false},
				},
			},
			expectedValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedValue, tt.schema.HasExternalFid())
		})
	}
}

func TestToTypeFormat(t *testing.T) {
	tests := []struct {
		name          string
		field         Field
		expectedValue TypeFormat
	}{
		{
			name: "integer field type",
			field: Field{
				Name:  "id",
				Type:  "integer",
				IsFid: true,
			},
			expectedValue: TypeFormat{
				Type:   "integer",
				Format: "",
			},
		},
		{
			name: "int field type",
			field: Field{
				Name: "count",
				Type: "int",
			},
			expectedValue: TypeFormat{
				Type:   "integer",
				Format: "",
			},
		},
		{
			name: "smallint field type",
			field: Field{
				Name: "smallcount",
				Type: "smallint",
			},
			expectedValue: TypeFormat{
				Type:   "integer",
				Format: "",
			},
		},
		{
			name: "bigint field type",
			field: Field{
				Name: "bigcount",
				Type: "bigint",
			},
			expectedValue: TypeFormat{
				Type:   "integer",
				Format: "",
			},
		},
		{
			name: "string field type",
			field: Field{
				Name: "name",
				Type: "string",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "geometry field type",
			field: Field{
				Name: "location",
				Type: "Point",
			},
			expectedValue: TypeFormat{
				Type:   "point",
				Format: "geometry-point",
			},
		},
		{
			name: "boolean field type",
			field: Field{
				Name: "active",
				Type: "boolean",
			},
			expectedValue: TypeFormat{
				Type:   "boolean",
				Format: "",
			},
		},
		{
			name: "bool field type (alternative for boolean)",
			field: Field{
				Name: "active",
				Type: "bool",
			},
			expectedValue: TypeFormat{
				Type:   "boolean",
				Format: "",
			},
		},
		{
			name: "text field type",
			field: Field{
				Name: "description",
				Type: "text",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "text with length field type",
			field: Field{
				Name: "description",
				Type: "text(33)",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "varchar with length field type",
			field: Field{
				Name: "description",
				Type: "varchar(33)",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "char field type",
			field: Field{
				Name: "code",
				Type: "char",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "varchar field type",
			field: Field{
				Name: "code",
				Type: "varchar",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
		{
			name: "float field type",
			field: Field{
				Name: "price",
				Type: "float",
			},
			expectedValue: TypeFormat{
				Type:   "number",
				Format: "double",
			},
		},
		{
			name: "double field type",
			field: Field{
				Name: "price",
				Type: "double",
			},
			expectedValue: TypeFormat{
				Type:   "number",
				Format: "double",
			},
		},
		{
			name: "numeric field type",
			field: Field{
				Name: "price",
				Type: "numeric",
			},
			expectedValue: TypeFormat{
				Type:   "number",
				Format: "double",
			},
		},
		{
			name: "uuid field type",
			field: Field{
				Name: "id",
				Type: "uuid",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "uuid",
			},
		},
		{
			name: "date field type",
			field: Field{
				Name: "start_date",
				Type: "date",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "date",
			},
		},
		{
			name: "time field type",
			field: Field{
				Name: "start_time",
				Type: "time",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "time",
			},
		},
		{
			name: "datetime field type",
			field: Field{
				Name: "created_at",
				Type: "datetime",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "date-time",
			},
		},
		{
			name: "timestamp field type",
			field: Field{
				Name: "updated_at",
				Type: "timestamp",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "date-time",
			},
		},
		{
			name: "geometry field type",
			field: Field{
				Name: "shape",
				Type: "geometry",
			},
			expectedValue: TypeFormat{
				Type:   "geometry",
				Format: "geometry-any",
			},
		},
		{
			name: "geometrycollection field type",
			field: Field{
				Name: "shapes",
				Type: "geometrycollection",
			},
			expectedValue: TypeFormat{
				Type:   "geometrycollection",
				Format: "geometry-any",
			},
		},
		{
			name: "linestring field type",
			field: Field{
				Name: "route",
				Type: "linestring",
			},
			expectedValue: TypeFormat{
				Type:   "linestring",
				Format: "geometry-linestring",
			},
		},
		{
			name: "polygon field type",
			field: Field{
				Name: "area",
				Type: "polygon",
			},
			expectedValue: TypeFormat{
				Type:   "polygon",
				Format: "geometry-polygon",
			},
		},
		{
			name: "multipoint field type",
			field: Field{
				Name: "points",
				Type: "multipoint",
			},
			expectedValue: TypeFormat{
				Type:   "multipoint",
				Format: "geometry-multipoint",
			},
		},
		{
			name: "multilinestring field type",
			field: Field{
				Name: "routes",
				Type: "multilinestring",
			},
			expectedValue: TypeFormat{
				Type:   "multilinestring",
				Format: "geometry-multilinestring",
			},
		},
		{
			name: "multipolygon field type",
			field: Field{
				Name: "areas",
				Type: "multipolygon",
			},
			expectedValue: TypeFormat{
				Type:   "multipolygon",
				Format: "geometry-multipolygon",
			},
		},
		{
			name: "mixed case type",
			field: Field{
				Name: "id",
				Type: "InTeGER",
			},
			expectedValue: TypeFormat{
				Type:   "integer",
				Format: "",
			},
		},
		{
			name: "unknown field type",
			field: Field{
				Name: "custom",
				Type: "custom_type",
			},
			expectedValue: TypeFormat{
				Type:   "string",
				Format: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedValue, tt.field.ToTypeFormat())
		})
	}
}
