// Licensed to the Apache Software Foundation (ASF) under one or more
// contributor license agreements.  See the NOTICE file distributed with
// this work for additional information regarding copyright ownership.
// The ASF licenses this file to You under the Apache License, Version 2.0
// (the "License"); you may not use this file except in compliance with
// the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ingress

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	ginkgo "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/apache/apisix-ingress-controller/test/e2e/scaffold"
)

var _ = ginkgo.Describe("suite-ingress-features: Status subresource Testing", func() {
	routeSuites := func(s *scaffold.Scaffold) {
		ginkgo.It("check the ApisixRoute status is recorded", func() {
			backendSvc, backendSvcPort := s.DefaultHTTPBackend()
			apisixRoute := fmt.Sprintf(`
apiVersion: apisix.apache.org/v2beta3
kind: ApisixRoute
metadata:
  name: httpbin-route
spec:
  http:
  - name: rule1
    match:
      hosts:
      - httpbin.com
      paths:
      - /ip
    backends:
    - serviceName: %s
      servicePort: %d
`, backendSvc, backendSvcPort[0])
			assert.Nil(ginkgo.GinkgoT(), s.CreateVersionedApisixResource(apisixRoute))

			err := s.EnsureNumApisixRoutesCreated(1)
			assert.Nil(ginkgo.GinkgoT(), err, "Checking number of routes")
			err = s.EnsureNumApisixUpstreamsCreated(1)
			assert.Nil(ginkgo.GinkgoT(), err, "Checking number of upstreams")
			// status should be recorded as successful
			output, err := s.GetOutputFromString("ar", "httpbin-route", "-o", "yaml")
			assert.Nil(ginkgo.GinkgoT(), err, "Get output of ApisixRoute resource")
			hasType := strings.Contains(output, "type: ResourcesAvailable")
			assert.True(ginkgo.GinkgoT(), hasType, "Status is recorded")
			hasMsg := strings.Contains(output, "message: Sync Successfully")
			assert.True(ginkgo.GinkgoT(), hasMsg, "Status is recorded")
		})
	}

	ginkgo.Describe("suite-ingress-features: ApisixRoute scaffold v2beta3", func() {
		routeSuites(scaffold.NewDefaultV2beta3Scaffold())
	})
	ginkgo.Describe("suite-ingress-features: ApisixRoute scaffold v2", func() {
		routeSuites(scaffold.NewDefaultV2Scaffold())
	})

	upSuite := func(s *scaffold.Scaffold) {
		ginkgo.It("check the ApisixUpstream status is recorded", func() {
			backendSvc, _ := s.DefaultHTTPBackend()
			apisixUpstream := fmt.Sprintf(`
apiVersion: apisix.apache.org/v2beta3
kind: ApisixUpstream
metadata:
  name: %s
spec:
  retries: 2
`, backendSvc)
			assert.Nil(ginkgo.GinkgoT(), s.CreateVersionedApisixResource(apisixUpstream))

			// status should be recorded as successful
			output, err := s.GetOutputFromString("au", backendSvc, "-o", "yaml")
			assert.Nil(ginkgo.GinkgoT(), err, "Get output of ApisixUpstream resource"+backendSvc)
			hasType := strings.Contains(output, "type: ResourcesAvailable")
			assert.True(ginkgo.GinkgoT(), hasType, "Status is recorded")
			hasMsg := strings.Contains(output, "message: Sync Successfully")
			assert.True(ginkgo.GinkgoT(), hasMsg, "Status is recorded")
		})
	}

	ginkgo.Describe("suite-ingress-features: ApisixUpstream scaffold v2beta3", func() {
		upSuite(scaffold.NewDefaultV2beta3Scaffold())
	})
	ginkgo.Describe("suite-ingress-features: ApisixUpstream scaffold v2", func() {
		upSuite(scaffold.NewDefaultV2Scaffold())
	})
})

var _ = ginkgo.Describe("suite-ingress-features: Ingress LB Status Testing", func() {
	opts := &scaffold.Options{
		Name:                  "default",
		IngressAPISIXReplicas: 1,
		APISIXPublishAddress:  "10.6.6.6",
	}
	s := scaffold.NewScaffold(opts)
	ginkgo.It("check the ingress lb status is updated", func() {
		backendSvc, backendPort := s.DefaultHTTPBackend()
		ing := fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: apisix
  name: ingress-v1-lb
spec:
  rules:
  - host: httpbin.org
    http:
      paths:
      - path: /ip
        pathType: Exact
        backend:
          service:
            name: %s
            port:
              number: %d
`, backendSvc, backendPort[0])
		err := s.CreateResourceFromString(ing)
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")
		time.Sleep(5 * time.Second)

		_ = s.NewAPISIXClient().GET("/ip").WithHeader("Host", "httpbin.org").Expect().Status(http.StatusOK)

		output, err := s.GetOutputFromString("ingress", "ingress-v1-lb", "-o", "jsonpath='{ .status.loadBalancer.ingress[0].ip }'")
		assert.Nil(ginkgo.GinkgoT(), err, "Get output of ingress status")

		hasIP := strings.Contains(output, "10.6.6.6")
		assert.True(ginkgo.GinkgoT(), hasIP, "LB Status is recorded")
	})
})

var _ = ginkgo.Describe("suite-ingress-features: disable status", func() {
	opts := &scaffold.Options{
		Name:                  "default",
		IngressAPISIXReplicas: 1,
		APISIXPublishAddress:  "10.6.6.6",
		DisableStatus:         true,
	}
	s := scaffold.NewScaffold(opts)
	ginkgo.It("check the ApisixRoute status is recorded", func() {
		backendSvc, backendSvcPort := s.DefaultHTTPBackend()
		apisixRoute := fmt.Sprintf(`
apiVersion: apisix.apache.org/v2
kind: ApisixRoute
metadata:
  name: httpbin-route
spec:
  http:
  - name: rule1
    match:
      hosts:
      - httpbin.com
      paths:
      - /ip
    backends:
    - serviceName: %s
      servicePort: %d
`, backendSvc, backendSvcPort[0])
		assert.Nil(ginkgo.GinkgoT(), s.CreateVersionedApisixResource(apisixRoute))

		err := s.EnsureNumApisixRoutesCreated(1)
		assert.Nil(ginkgo.GinkgoT(), err, "Checking number of routes")
		err = s.EnsureNumApisixUpstreamsCreated(1)
		assert.Nil(ginkgo.GinkgoT(), err, "Checking number of upstreams")
		// status should be recorded as successful
		output, err := s.GetOutputFromString("ar", "httpbin-route", "-o", "jsonpath='{ .status }'")
		assert.Nil(ginkgo.GinkgoT(), err)
		assert.Equal(ginkgo.GinkgoT(), "''", output)
	})

	ginkgo.It("check the ingress lb status is updated", func() {
		backendSvc, backendPort := s.DefaultHTTPBackend()
		ing := fmt.Sprintf(`
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    kubernetes.io/ingress.class: apisix
  name: ingress-v1-lb
spec:
  rules:
  - host: httpbin.org
    http:
      paths:
      - path: /ip
        pathType: Exact
        backend:
          service:
            name: %s
            port:
              number: %d
`, backendSvc, backendPort[0])
		err := s.CreateResourceFromString(ing)
		assert.Nil(ginkgo.GinkgoT(), err, "creating ingress")
		time.Sleep(5 * time.Second)

		_ = s.NewAPISIXClient().GET("/ip").WithHeader("Host", "httpbin.org").Expect().Status(http.StatusOK)

		output, err := s.GetOutputFromString("ingress", "ingress-v1-lb", "-o", "jsonpath='{ .status.loadBalancer }'")
		assert.Nil(ginkgo.GinkgoT(), err, "Get output of ingress status")

		assert.Equal(ginkgo.GinkgoT(), "'{}'", output)
	})
})
