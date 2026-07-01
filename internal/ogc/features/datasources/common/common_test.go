package common

import (
	"testing"
	"time"

	"github.com/PDOK/gokoala/internal/ogc/features/datasources"
	"github.com/PDOK/gokoala/internal/ogc/features/domain"
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
	start, _ := time.Parse(time.RFC3339, "2023-11-10T23:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2023-11-15T23:00:00Z")

	testCases := []struct {
		name         string
		criteria     datasources.TemporalCriteria
		symbol       string
		expectedSQL  string
		expectedArgs map[string]any
	}{
		{
			name: "Instant",
			criteria: datasources.TemporalCriteria{
				DateTime:          domain.DateTime{Instant: &now},
				StartDateProperty: "start_date",
				EndDateProperty:   "end_date",
			},
			symbol:      ":",
			expectedSQL: ` and "start_date" <= :instant and ("end_date" >= :instant or "end_date" is null)`,
			expectedArgs: map[string]any{
				"instant": &now,
			},
		},
		{
			name: "Closed interval",
			criteria: datasources.TemporalCriteria{
				DateTime:          domain.DateTime{IntervalStart: &start, IntervalEnd: &end},
				StartDateProperty: "start_date",
				EndDateProperty:   "end_date",
			},
			symbol:      ":",
			expectedSQL: ` and ("start_date" <= :intervalEnd or "start_date" is null) and ("end_date" >= :intervalStart or "end_date" is null)`,
			expectedArgs: map[string]any{
				"intervalStart": &start,
				"intervalEnd":   &end,
			},
		},
		{
			name: "Open-start interval",
			criteria: datasources.TemporalCriteria{
				DateTime:          domain.DateTime{IntervalStart: nil, IntervalEnd: &end},
				StartDateProperty: "start_date",
				EndDateProperty:   "end_date",
			},
			symbol:      ":",
			expectedSQL: ` and "start_date" <= :intervalEnd or "start_date" is null`,
			expectedArgs: map[string]any{
				"intervalEnd": &end,
			},
		},
		{
			name: "Open-end interval",
			criteria: datasources.TemporalCriteria{
				DateTime:          domain.DateTime{IntervalStart: &start, IntervalEnd: nil},
				StartDateProperty: "start_date",
				EndDateProperty:   "end_date",
			},
			symbol:      ":",
			expectedSQL: ` and "end_date" >= :intervalStart or "end_date" is null`,
			expectedArgs: map[string]any{
				"intervalStart": &start,
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
			result := ColumnsToSQL(tt.columns, true)
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
				"key1": {Name: "table1"},
				"key2": {Name: "table2"},
			},
			expected: 0, // No warnings expected
		},
		{
			name: "Duplicate tables",
			input: map[string]*Table{
				"key1": {Name: "table1"},
				"key2": {Name: "table1"},
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
