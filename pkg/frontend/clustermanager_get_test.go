package frontend

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/ARO-RP/pkg/api"
	"github.com/Azure/ARO-RP/pkg/metrics/noop"
	testdatabase "github.com/Azure/ARO-RP/test/database"
)

// Copyright (c) Microsoft Corporation.
// Licensed under the Apache License 2.0.

func TestGetClusterManagerConfiguration(t *testing.T) {
	ctx := context.Background()

	mockSubscriptionId := "00000000-0000-0000-0000-000000000000"
	mockResourceGroupName := "resourceGroup"
	mockResourceType := "openShiftClusters"
	mockOcmResourceType := "syncSet"
	mockOcmResourceName := "mySyncSet"
	apiVersion := "2022-09-04"

	type test struct {
		name           string
		fixture        func(f *testdatabase.Fixture)
		wantStatusCode int
		wantResponse   *api.ClusterManagerConfigurationDocuments
		wantError      string
	}
	for _, tt := range []*test{
		{
			name: "single syncset",
			fixture: func(f *testdatabase.Fixture) {
				f.AddClusterManagerConfigurationDocuments(
					&api.ClusterManagerConfigurationDocument{
						ID:           mockSubscriptionId,
						Key:          "/subscriptions/subscriptionid/resourcegroups/resourcegroup/providers/microsoft.redhatopenshift/openshiftclusters/resourcename/syncSets/mySyncSet",
						ResourceID:   "",
						PartitionKey: "",
						ClusterManagerConfiguration: &api.ClusterManagerConfiguration{
							Name: "myCluster",
						},
						SyncSet: &api.SyncSet{
							Name: "mySyncSet",
							ID:   "/subscriptions/subscriptionId/resourceGroups/resourceGroup/providers/Microsoft.RedHatOpenShift/OpenShiftClusters/resourceName/syncSets/mySyncSet",
							Type: "Microsoft.RedHatOpenShift/OpenShiftClusters/SyncSets",
							Properties: api.SyncSetProperties{
								Resources: "eyAKICAiYXBpVmVyc2lvbiI6ICJoaXZlLm9wZW5zaGlmdC5pby92MSIsCiAgImtpbmQiOiAiU3luY1NldCIsCiAgIm1ldGFkYXRhIjogewogICAgIm5hbWUiOiAic2FtcGxlIiwKICAgICJuYW1lc3BhY2UiOiAiYXJvLWY2MGFlOGEyLWJjYTEtNDk4Ny05MDU2LWYyZjZhMTgzN2NhYSIKICB9LAogICJzcGVjIjogewogICAgImNsdXN0ZXJEZXBsb3ltZW50UmVmcyI6IFtdLAogICAgInJlc291cmNlcyI6IFsKICAgICAgewogICAgICAgICJhcGlWZXJzaW9uIjogInYxIiwKICAgICAgICAia2luZCI6ICJDb25maWdNYXAiLAogICAgICAgICJtZXRhZGF0YSI6IHsKICAgICAgICAgICJuYW1lIjogIm15Y29uZmlnbWFwIgogICAgICAgIH0KICAgICAgfQogICAgXQogIH0KfQo=",
							},
						},
						CorrelationData: &api.CorrelationData{},
					},
				)
			},
			wantStatusCode: http.StatusOK,
			wantResponse:   &api.ClusterManagerConfigurationDocuments{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ti := newTestInfra(t).WithClusterManagerConfigurations()
			defer ti.done()

			err := ti.buildFixtures(tt.fixture)
			if err != nil {
				t.Fatal(err)
			}

			f, err := NewFrontend(ctx, ti.audit, ti.log, ti.env, nil, ti.clusterManagerDatabase, nil, nil, nil, api.APIs, &noop.Noop{}, nil, nil, nil, nil)

			if err != nil {
				t.Fatal(err)
			}

			go f.Run(ctx, nil, nil)

			resp, b, err := ti.request(http.MethodGet,
				// "/subscriptions/{subscriptionId}/resourcegroups/{resourceGroupName}/providers/{resourceProviderNamespace}/{resourceType}
				// /{resourceName}/{ocmResourceType}/{ocmResourceName}"
				fmt.Sprintf("https://server/subscriptions/%s/resourcegroups/%s/providers/Microsoft.RedHatOpenShift/%s/%s/%s/%s?api-version=%s",
					mockSubscriptionId,
					mockResourceGroupName,
					mockResourceType,
					mockResourceGroupName,
					mockOcmResourceType,
					mockOcmResourceName,
					apiVersion,
				),
				nil, nil)
			if err != nil {
				t.Fatalf("%s: %s", err, string(b))
			}

			err = validateResponse(resp, b, tt.wantStatusCode, tt.wantError, tt.wantResponse)
			if err != nil {
				t.Errorf("%s: %s", err, string(b))
			}
		})
	}
}
