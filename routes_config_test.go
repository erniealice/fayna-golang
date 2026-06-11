package fayna_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	fulfillmentdomain "github.com/erniealice/fayna-golang/domain/fulfillment"
	operation "github.com/erniealice/fayna-golang/domain/operation"
)

type routeContractCase struct {
	name         string
	routes       any
	routeMap     map[string]string
	unmappedURLs map[string]bool
}

func TestDefaultRoutes_AllStringFieldsNonEmpty(t *testing.T) {
	t.Parallel()

	for _, tc := range faynaRouteContractCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertAllStringFieldsNonEmpty(t, tc.routes)
		})
	}
}

func TestRouteMap_ValuesBelongToStructAndCoverRouteFields(t *testing.T) {
	t.Parallel()

	for _, tc := range faynaRouteContractCases() {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assertRouteMapContract(t, tc.routes, tc.routeMap, tc.unmappedURLs)
		})
	}
}

func faynaRouteContractCases() []routeContractCase {
	return []routeContractCase{
		{
			name:     "JobRoutes",
			routes:   operation.DefaultJobRoutes(),
			routeMap: operation.DefaultJobRoutes().RouteMap(),
		},
		{
			name:     "JobTemplateRoutes",
			routes:   operation.DefaultJobTemplateRoutes(),
			routeMap: operation.DefaultJobTemplateRoutes().RouteMap(),
		},
		{
			name:     "JobActivityRoutes",
			routes:   operation.DefaultJobActivityRoutes(),
			routeMap: operation.DefaultJobActivityRoutes().RouteMap(),
		},
		{
			name:     "OutcomeCriteriaRoutes",
			routes:   operation.DefaultOutcomeCriteriaRoutes(),
			routeMap: operation.DefaultOutcomeCriteriaRoutes().RouteMap(),
		},
		{
			name:     "TaskOutcomeRoutes",
			routes:   operation.DefaultTaskOutcomeRoutes(),
			routeMap: operation.DefaultTaskOutcomeRoutes().RouteMap(),
		},
		{
			name:     "OutcomeSummaryRoutes",
			routes:   operation.DefaultOutcomeSummaryRoutes(),
			routeMap: operation.DefaultOutcomeSummaryRoutes().RouteMap(),
		},
		{
			name:     "FulfillmentRoutes",
			routes:   fulfillmentdomain.DefaultFulfillmentRoutes(),
			routeMap: fulfillmentdomain.DefaultFulfillmentRoutes().RouteMap(),
		},
		{
			name:     "JobTemplatePhaseRoutes",
			routes:   operation.DefaultJobTemplatePhaseRoutes(),
			routeMap: operation.DefaultJobTemplatePhaseRoutes().RouteMap(),
		},
		{
			name:     "JobTemplateTaskRoutes",
			routes:   operation.DefaultJobTemplateTaskRoutes(),
			routeMap: operation.DefaultJobTemplateTaskRoutes().RouteMap(),
		},
	}
}

func assertAllStringFieldsNonEmpty(t *testing.T, routes any) {
	t.Helper()

	value := reflect.ValueOf(routes)
	typ := value.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.String {
			continue
		}
		if value.Field(i).String() == "" {
			t.Fatalf("%s.%s should not be empty", typ.Name(), field.Name)
		}
	}
}

func assertRouteMapContract(t *testing.T, routes any, routeMap map[string]string, unmappedURLs map[string]bool) {
	t.Helper()

	routeFields := collectURLFields(routes)
	var missing []string

	for key, value := range routeMap {
		if key == "" {
			t.Fatalf("%T RouteMap contains an empty key", routes)
		}
		if value == "" {
			t.Fatalf("%T RouteMap[%q] should not be empty", routes, key)
		}
		if !containsValue(routeFields, value) {
			t.Fatalf("%T RouteMap[%q]=%q does not match any URL field", routes, key, value)
		}
	}

	for fieldName, value := range routeFields {
		if unmappedURLs[fieldName] {
			continue
		}
		if !containsMapValue(routeMap, value) {
			missing = append(missing, fieldName)
		}
	}

	if len(missing) > 0 {
		slices.Sort(missing)
		t.Fatalf("%T RouteMap is missing URL fields: %s", routes, strings.Join(missing, ", "))
	}
}

func collectURLFields(routes any) map[string]string {
	value := reflect.ValueOf(routes)
	typ := value.Type()
	fields := make(map[string]string)

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if field.Type.Kind() != reflect.String {
			continue
		}
		if !strings.HasSuffix(field.Name, "URL") {
			continue
		}
		fields[field.Name] = value.Field(i).String()
	}

	return fields
}

func containsValue(values map[string]string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsMapValue(values map[string]string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
