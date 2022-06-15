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

package config

import "time"

var MyEnvConfig EnvConfig

type EnvConfig struct {

	// Application provides the application configurations.
	Application struct {
		Debug          bool   `env:"DEBUG"            envDefault:"false"`
		UseKubeCfg     bool   `env:"USE_KUBECONFIG"   envDefault:"false"`
		KubeConfigFile string `env:"KUBECONFIG"       envDefault:"${HOME}/.kube/config" envExpand:"true"`
	}

	// HTTPServer provides the HTTP server configuration.
	HTTPServer struct {
		Port string `env:"PORT"     envDefault:"8080"`

		ReadTimeout          time.Duration `env:"APP_READ_TIMEOUT" envDefault:"5s"`
		WriteTimeout         time.Duration `env:"APP_WRITE_TIMEOUT" envDefault:"5s"`
		MaxConnsPerIP        int           `env:"APP_MAX_CONN_PER_IP" envDefault:"50"`
		MaxRequestsPerConn   int           `env:"APP_MAX_REQUESTS_PER_CONN" envDefault:"10"`
		MaxKeepaliveDuration time.Duration `env:"APP_MAX_KEEP_ALIVE_DURATION" envDefault:"5s"`
	}
}
