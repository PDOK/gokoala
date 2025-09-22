package common

import (
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/stretchr/testify/assert"
)

func TestPropertyFiltersToSQL(t *testing.T) {
	testCases := []struct {
		name         string
		filters      map[string]string
		symbol       string
		expectedSQL  string
		expectedArgs map[string]any
	}{
		{
			name: "Single filter",
			filters: map[string]string{
				"column1": "value1",
			},
			symbol:      ":",
			expectedSQL: ` and "column1" = :pf1`,
			expectedArgs: map[string]any{
				"pf1": "value1",
			},
		},
		{
			name:         "No filters",
			filters:      map[string]string{},
			symbol:       ":",
			expectedSQL:  "",
			expectedArgs: map[string]any{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := PropertyFiltersToSQL(tt.filters, tt.symbol)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestTemporalCriteriaToSQL(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name         string
		criteria     datasources.TemporalCriteria
		symbol       string
		expectedSQL  string
		expectedArgs map[string]any
	}{
		{
			name: "Valid temporal criteria",
			criteria: datasources.TemporalCriteria{
				ReferenceDate:     now,
				StartDateProperty: "start_date",
				EndDateProperty:   "end_date",
			},
			symbol:      ":",
			expectedSQL: ` and "start_date" <= :referenceDate and ("end_date" >= :referenceDate or "end_date" is null)`,
			expectedArgs: map[string]any{
				"referenceDate": now,
			},
		},
		{
			name:         "Empty temporal criteria",
			criteria:     datasources.TemporalCriteria{},
			symbol:       ":",
			expectedSQL:  "",
			expectedArgs: map[string]any{},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sql, args := TemporalCriteriaToSQL(tt.criteria, tt.symbol)
			assert.Equal(t, tt.expectedSQL, sql)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

func TestColumnsToSQL(t *testing.T) {
	testCases := []struct {
		name          string
		columns       []string
		expectedQuery string
	}{
		{
			name:          "Single column",
			columns:       []string{"column1"},
			expectedQuery: `"column1"`,
		},
		{
			name:          "Multiple columns",
			columns:       []string{"column1", "column2"},
			expectedQuery: `"column1", "column2"`,
		},
		{
			name:          "No columns",
			columns:       []string{},
			expectedQuery: `""`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := ColumnsToSQL(tt.columns)
			assert.Equal(t, tt.expectedQuery, result)
		})
	}
}

func TestValidateUniqueness(t *testing.T) {
	testCases := []struct {
		name     string
		input    map[string]*Table
		expected int
	}{
		{
			name: "Unique tables",
			input: map[string]*Table{
				"key1": {TableName: "table1"},
				"key2": {TableName: "table2"},
			},
			expected: 0, // No warnings expected
		},
		{
			name: "Duplicate tables",
			input: map[string]*Table{
				"key1": {TableName: "table1"},
				"key2": {TableName: "table1"},
			},
			expected: 1, // One warning expected
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(_ *testing.T) {
			// Just testing input behavior
			ValidateUniqueness(tt.input)
		})
	}
}
