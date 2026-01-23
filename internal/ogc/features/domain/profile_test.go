package domain

import (
	neturl "net/url"
	"testing"

	"github.com/PDOK/gokoala/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapRelationUsingProfile(t *testing.T) {
	tests := []struct {
		name                       string
		profile                    ProfileName
		columnName                 string
		columnValue                any
		externalFidCol             string
		expectedColName            string
		expectedColNameUnformatted string
		expectedColVal             any
	}{
		{
			name:                       "RelAsLink",
			profile:                    RelAsLink,
			columnName:                 "another_collection_external_fid",
			columnValue:                "123",
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection.href",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             "http://example.com/collections/another_collection/items/123",
		},
		{
			name:                       "RelAsLink with nil value",
			profile:                    RelAsLink,
			columnName:                 "another_collection_external_fid",
			columnValue:                nil,
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection.href",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             nil,
		},
		{
			name:                       "RelAsLink with infix in column name",
			profile:                    RelAsLink,
			columnName:                 "another_collection_some_infix_external_fid",
			columnValue:                "123",
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection_some_infix.href",
			expectedColNameUnformatted: "another_collection_some_infix",
			expectedColVal:             "http://example.com/collections/another_collection/items/123",
		},
		{
			name:                       "RelAsLink with similar collection name (make sure exact match is selected)",
			profile:                    RelAsLink,
			columnName:                 "baz_bazoo_boo_external_fid",
			columnValue:                "123",
			externalFidCol:             "external_fid",
			expectedColName:            "baz_bazoo_boo.href",
			expectedColNameUnformatted: "baz_bazoo_boo",
			expectedColVal:             "http://example.com/collections/baz_bazoo_boo/items/123",
		},
		{
			name:                       "RelAsKey",
			profile:                    RelAsKey,
			columnName:                 "another_collection_external_fid",
			columnValue:                "123",
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             "123",
		},
		{
			name:                       "RelAsKey with nil value",
			profile:                    RelAsKey,
			columnName:                 "another_collection_external_fid",
			columnValue:                nil,
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             nil,
		},
		{
			name:                       "RelAsURI",
			profile:                    RelAsURI,
			columnName:                 "another_collection_external_fid",
			columnValue:                "123",
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             "http://example.com/collections/another_collection/items/123",
		},
		{
			name:                       "RelAsURI with nil value",
			profile:                    RelAsURI,
			columnName:                 "another_collection_external_fid",
			columnValue:                nil,
			externalFidCol:             "external_fid",
			expectedColName:            "another_collection",
			expectedColNameUnformatted: "another_collection",
			expectedColVal:             nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := neturl.Parse("http://example.com")
			require.NoError(t, err)
			collections := config.CollectionsFeatures{
				config.CollectionFeatures{
					ID: "some_collection",
				}, config.CollectionFeatures{
					ID: "another_collection",
				}, config.CollectionFeatures{
					ID: "foo",
				}, config.CollectionFeatures{
					ID: "bar",
				}, config.CollectionFeatures{
					ID: "baz_bazoo",
				}, config.CollectionFeatures{
					ID: "baz_bazoo_boo",
				}, config.CollectionFeatures{
					ID: "baz_bazoo_boo_foo",
				},
			}
			schema, err := NewSchema([]Field{
				{
					Name:            tt.columnName,
					Type:            "string",
					FeatureRelation: NewFeatureRelation("some table", tt.columnName, tt.externalFidCol, collections),
				}}, "fid", "")
			require.NoError(t, err)
			profile := NewProfile(tt.profile, *url, *schema)
			newColName, newColNameUnformatted, newColVal := profile.MapRelationUsingProfile(tt.columnName, tt.columnValue, tt.externalFidCol)
			assert.Equal(t, tt.expectedColName, newColName)
			assert.Equal(t, tt.expectedColNameUnformatted, newColNameUnformatted)
			assert.Equal(t, tt.expectedColVal, newColVal)
		})
	}
}
