package domain

import (
	neturl "net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapRelationUsingProfile(t *testing.T) {
	tests := []struct {
		name            string
		profile         ProfileName
		columnName      string
		columnValue     any
		externalFidCol  string
		expectedColName string
		expectedColVal  any
	}{
		{
			name:            "RelAsLink",
			profile:         RelAsLink,
			columnName:      "another_collection_external_fid",
			columnValue:     "123",
			externalFidCol:  "external_fid",
			expectedColName: "another_collection.href",
			expectedColVal:  "http://example.com/collections/another_collection/items/123",
		},
		{
			name:            "RelAsLink with nil value",
			profile:         RelAsLink,
			columnName:      "another_collection_external_fid",
			columnValue:     nil,
			externalFidCol:  "external_fid",
			expectedColName: "another_collection.href",
			expectedColVal:  nil,
		},
		{
			name:            "RelAsKey",
			profile:         RelAsKey,
			columnName:      "another_collection_external_fid",
			columnValue:     "123",
			externalFidCol:  "external_fid",
			expectedColName: "another_collection",
			expectedColVal:  "123",
		},
		{
			name:            "RelAsKey with nil value",
			profile:         RelAsKey,
			columnName:      "another_collection_external_fid",
			columnValue:     nil,
			externalFidCol:  "external_fid",
			expectedColName: "another_collection",
			expectedColVal:  nil,
		},
		{
			name:            "RelAsURI",
			profile:         RelAsURI,
			columnName:      "another_collection_external_fid",
			columnValue:     "123",
			externalFidCol:  "external_fid",
			expectedColName: "another_collection",
			expectedColVal:  "http://example.com/collections/another_collection/items/123",
		},
		{
			name:            "RelAsURI with nil value",
			profile:         RelAsURI,
			columnName:      "another_collection_external_fid",
			columnValue:     nil,
			externalFidCol:  "external_fid",
			expectedColName: "another_collection",
			expectedColVal:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, err := neturl.Parse("http://example.com")
			assert.NoError(t, err)
			profile := NewProfile(tt.profile, *url)
			newColName, newColVal := profile.MapRelationUsingProfile(tt.columnName, tt.columnValue, tt.externalFidCol)
			assert.Equal(t, tt.expectedColName, newColName)
			assert.Equal(t, tt.expectedColVal, newColVal)
		})
	}
}
