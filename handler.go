// Copyright 2019 Muhammet Arslan <github.com/geass>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"git2.gnt-global.com/jlab/gdeploy/domains-viewers/config"
	"git2.gnt-global.com/jlab/gdeploy/domains-viewers/pkg/k8s"
	"git2.gnt-global.com/jlab/gdeploy/domains-viewers/pkg/logger"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// Handler exposes the Handler methods
type Handler struct{}

// type Domains map[string]string

type InternalDomains struct {
	Cname map[string]string   `json:"cname"`
	Svc   map[string][]string `json:"svc"`
}

type Domains struct {
	External map[string]string `json:"external"`
	Internal InternalDomains   `json:"internal"`
}

// Index function renders the dashboard index page
func (h *Handler) Index() fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		logger := logger.New("domains-viewers", config.MyEnvConfig.Application.Debug)

		domains := Domains{}

		// Debug Log
		logger.Debug(domains)
		clientSet := k8s.GetClientSet(logger)
		dynamicClientSet := k8s.GetDynamic(logger)
		ctx.SetContentType("application/json; charset=utf-8")
		GetExternalServiceDomains(clientSet, logger, &domains.External)
		GetVirtualServiceDomains(dynamicClientSet, logger, &domains.External)
		GetInternalServiceDomains(clientSet, logger, &domains.Internal)
		GetKnativeServiceDomains(dynamicClientSet, logger, &domains)
		// fmt.Fprintf(ctx, `{"ingresses":"%s","internal":"%s","knative":"%s"}`, GetExternalServiceDomains(clientSet, logger), GetInternalServiceDomains(clientSet, logger), GetKnativeServiceDomains(clientSet, logger))
		bytes, err := json.Marshal(domains)
		if err != nil {
			logger.Panicw("error when marshaling domains", "domains", domains)
		}
		fmt.Fprint(ctx, string(bytes))
	}
}

func GetExternalServiceDomains(clientSet kubernetes.Interface, logger *zap.SugaredLogger, domains *map[string]string) {
	if *domains == nil {
		logger.Debug("nil domains in GetExternalServiceDomains")
		*domains = make(map[string]string)
	}
	ingList, err := clientSet.NetworkingV1().Ingresses("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		logger.Panicw("error when getting ingress details", "ingress List", ingList, "err", err)
	}
	for _, ing := range ingList.Items {
		for _, rule := range ing.Spec.Rules {
			for _, path := range rule.HTTP.Paths {
				subPath := ""
				if len(path.Path) > 1 {
					subPath = path.Path
					if !strings.HasSuffix(path.Path, "/") {
						subPath += "/"
					}
				}
				(*domains)[rule.Host+subPath] = fmt.Sprintf("%s.%s:%d", path.Backend.Service.Name, ing.Namespace, path.Backend.Service.Port.Number)
			}
		}
	}
}
func GetVirtualServiceDomains(clientSet dynamic.Interface, logger *zap.SugaredLogger, domains *map[string]string) {
	if *domains == nil {
		logger.Debug("nil domains in GetVirtualServiceDomains")
		*domains = make(map[string]string)
	}
	objs, err := clientSet.Resource(virtualServiceResource).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		if err.Error() == "the server could not find the requested resource" {
			return
		} else {
			logger.Panicw("error when getting virtualservice details", "objs", objs, "err", err)
		}
	}

	for _, item := range objs.Items {
		gateways, found, err := unstructured.NestedStringSlice(item.UnstructuredContent(), "spec", "gateways")
		if err != nil {
			logger.Panicw("error when getting virtualservice spec.gateways", "objs", objs, "err", err)
		}
		if !found {
			return
		}
		internal := false
		for _, gw := range gateways {
			if strings.Contains(gw, "knative") || strings.Contains(gw, "mesh") {
				internal = true
			}
		}
		if internal {
			return
		}
		hosts, found, err := unstructured.NestedStringSlice(item.UnstructuredContent(), "spec", "hosts")
		if err != nil {
			logger.Panicw("error when getting virtualservice spec.hosts", "objs", objs, "err", err)
		}
		if !found {
			return
		}
		backends, found, err := unstructured.NestedSlice(item.UnstructuredContent(), "spec", "http")
		if err != nil {
			logger.Panicw("error when getting virtualservice spec.http", "objs", objs, "err", err)
		}
		if !found {
			return
		}

		for _, backend := range backends {
			backendMap := backend.(map[string]interface{})
			for _, host := range hosts {
				// matches := []string{"/"}
				// if val, ok := backend["match"]; ok {

				// }
				logger.Info(host)
				if routes, ok := backendMap["route"]; ok {
					for _, route := range routes.([]interface{}) {
						if _, ok := route.(map[string]interface{})["destination"]; ok {
							// if err != nil {
							// 	logger.Panicw("error when getting virtualservice spec.http", "objs", objs, "err", err)
							// }
							// if !found {
							// 	return
							// }

							// if match in backend, else match /
							// if destination in backend, else return
							// key = host, value = destination.host:destination.port.number
							(*domains)[host] = "abc"
						}
					}
				}
			}
		}
	}
}

func GetInternalServiceDomains(clientSet kubernetes.Interface, logger *zap.SugaredLogger, domains *InternalDomains) {
	if domains.Cname == nil {
		domains.Cname = make(map[string]string)
	}
	if domains.Svc == nil {
		domains.Svc = make(map[string][]string)
	}
	svcList, err := clientSet.CoreV1().Services("").List(context.TODO(), v1.ListOptions{})
	if err != nil {
		logger.Panicw("error when getting internal services details", "service List", svcList, "err", err)
	}
	logger.Debug(len(svcList.Items))
	for _, svc := range svcList.Items {
		if svc.Spec.Type == "ExternalName" {
			domains.Cname[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)] = strings.Replace(svc.Spec.ExternalName, ".svc.cluster.local", "", 1)
		} else {
			portStrings := []string{}
			for _, port := range svc.Spec.Ports {

				targetPort := ""

				switch port.TargetPort.Type {
				case intstr.Int:
					targetPort = fmt.Sprint(port.TargetPort.IntVal)
				case intstr.String:
					targetPort = port.TargetPort.StrVal
				}
				if svc.Spec.Type == "NodePort" {
					portStrings = append(portStrings, fmt.Sprintf("%d:%s:%d/%v", port.Port, targetPort, port.NodePort, port.Protocol))
				} else {
					portStrings = append(portStrings, fmt.Sprintf("%d:%s/%v", port.Port, targetPort, port.Protocol))
				}
			}
			domains.Svc[fmt.Sprintf("%s.%s", svc.Name, svc.Namespace)] = portStrings
		}
	}
}

var knativeServiceResource = schema.GroupVersionResource{
	Group:    "serving.knative.dev",
	Version:  "v1",
	Resource: "services",
}

var virtualServiceResource = schema.GroupVersionResource{
	Group:    "networking.istio.io",
	Version:  "v1beta1",
	Resource: "virtualservices",
}

func GetKnativeServiceDomains(clientSet dynamic.Interface, logger *zap.SugaredLogger, domains *Domains) {
	if domains.External == nil {
		domains.External = make(map[string]string)
	}
	if domains.Internal.Cname == nil {
		domains.Internal.Cname = make(map[string]string)
	}
	if domains.Internal.Svc == nil {
		domains.Internal.Svc = make(map[string][]string)
	}
	objs, err := clientSet.Resource(knativeServiceResource).List(context.TODO(), v1.ListOptions{
		LabelSelector: "route.external=true",
	})
	if err != nil {
		if err.Error() == "the server could not find the requested resource" {
			return
		} else {
			logger.Panicw("error when getting knative details", "objs", objs, "err", err)
		}
	}

	for _, item := range objs.Items {
		internalUrl := fmt.Sprintf("%s.%s", item.GetName(), item.GetNamespace())
		domains.Internal.Svc[internalUrl] = []string{
			"80:knative-container-port/HTTP",
		}
		externalUrl, found, err := unstructured.NestedString(item.UnstructuredContent(), "status", "url")
		if err != nil {
			logger.Panicw("error when getting knative ksvc.status.url", "objs", objs, "err", err)
		}
		if found {
			domains.External[strings.Replace(externalUrl, "https://", "", 1)] = internalUrl
		}
	}
}
