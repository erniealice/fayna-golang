package outcome_summary

import (
	"reflect"
	"testing"
)

func TestBuildClientReportPhaseCatalogFiltersDeduplicatesAndOrders(t *testing.T) {
	entries := []ClientReportPhaseEntry{
		{Code: "progress_report", Name: "Progress Report", Order: 3, Active: true},
		{Code: "s2", Name: "Semester 2", Order: 2, Active: true},
		{Code: "S1", Name: "Semester 1", Order: 1, Active: true},
		{Code: " s1 ", Name: "Semester One", Order: 1, Active: true},
		{Code: "unused", Name: "Unused", Order: 0, Active: false},
		{Code: "  ", Name: "Blank code", Order: 0, Active: true},
		{Code: "quarter", Name: "Quarter", Order: 2, Active: true},
	}

	got := BuildClientReportPhaseCatalog(entries)
	want := []ClientReportPhase{
		{Code: "S1", Name: "Semester 1", Order: 1},
		{Code: "quarter", Name: "Quarter", Order: 2},
		{Code: "s2", Name: "Semester 2", Order: 2},
		{Code: "progress_report", Name: "Progress Report", Order: 3},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildClientReportPhaseCatalog() = %#v, want %#v", got, want)
	}
}

func TestBuildClientReportPhaseCatalogUsesCodeWhenNameIsBlank(t *testing.T) {
	got := BuildClientReportPhaseCatalog([]ClientReportPhaseEntry{{Code: "s1", Active: true}})
	want := []ClientReportPhase{{Code: "s1", Name: "s1", Order: 0}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("BuildClientReportPhaseCatalog() = %#v, want %#v", got, want)
	}
}
