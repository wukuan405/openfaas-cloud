package test

import (
	"testing"
)

func Test_CoreEdgeAuthDep(t *testing.T) {
	var parts []string
	want := makeEdgeAuthDep(true)
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/ofc-core/edge-auth-dep.yaml", want, t)
}

func Test_CoreEdgeAuthDep_NonHttpProbe(t *testing.T) {
	parts := []string{
		"--set", "global.httpProbe=false",
	}
	want := makeEdgeAuthDep(false)
	runYamlTest(parts,"./tmp/openfaas-cloud/templates/ofc-core/edge-auth-dep.yaml", want, t)

}



func makeEdgeAuthDep(httpProbe bool) YamlSpec {
	labels := make(map[string]string)

	labels["app.kubernetes.io/component"] = "edge-auth"
	labels["app.kubernetes.io/instance"] = "RELEASE-NAME"
	labels["app.kubernetes.io/managed-by"] = "Helm"
	labels["app.kubernetes.io/name"] = "openfaas-cloud"
	labels["helm.sh/chart"] = "openfaas-cloud-0.11.9"


	deployVolumes := makeDeployVolumes([]string{"jwt-private-key", "jwt-public-key", "of-client-secret"})
	containerVolumes := makeContainerVolumes()
	containerEnvironment := makeContainerEnv()

	var readinessProbe ReadinessProbe
	if httpProbe {
		readinessProbe = ReadinessProbe{
			HttpGet: HttpProbe{
				Path: "/healthz",
				Port: 8080,
			},
			TimeoutSeconds: 5,
		}
	} else {
		readinessProbe = ReadinessProbe{
			ExecProbe:      ExecProbe{
				Command: []string{"wget","--quiet", "--tries=1", "--timeout=5", "--spider", "http://localhost:8080/healthz"},
			},
			TimeoutSeconds: 5,
		}
	}
	return YamlSpec{
		ApiVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata: MetadataItems{
			Name:        "edge-auth",
			Labels:      labels,
		},
		Spec: Spec{
			Replicas:    1,
			Selector:    MatchLabelSelector{MatchLabels: map[string]string{"app": "edge-auth"}},
			Template: SpecTemplate{
				Metadata: MetadataItems{
					Annotations: map[string]string{"prometheus.io.scrape":"false"},
					Labels: map[string]string{"app":"edge-auth"},
				},
				Spec: TemplateSpec{
					Volumes: deployVolumes,
					Containers: []DeploymentContainers{{
						Name:                    "edge-auth",
						Image:                   "openfaas/edge-auth:0.6.3",
						ImagePullPolicy:         "IfNotPresent",
						ContainerReadinessProbe: readinessProbe,
						ContainerEnvironment: containerEnvironment,
						Ports:                   []ContainerPort{{
							Port:     8080,
							Protocol: "TCP",
						}},
						Volumes:                 containerVolumes,
					}},
				},
			},
		},
	}
}

func makeContainerVolumes() []ContainerVolume {
	var vols []ContainerVolume
	vols = append(vols, ContainerVolume{
		Name:      "jwt-private-key",
		ReadOnly:  true,
		MountPath: "/var/secrets/private/",
	})

	vols = append(vols, ContainerVolume{
		Name:      "jwt-public-key",
		ReadOnly:  true,
		MountPath: "/var/secrets/public",
	})

	vols = append(vols, ContainerVolume{
		Name:      "of-client-secret",
		ReadOnly:  true,
		MountPath: "/var/secrets/of-client-secret",
	})

	return vols
}

func makeContainerEnv() []Environment {
	var environ []Environment
	environ = append(environ, Environment{ Name:  "port", Value: "8080"})
	environ = append(environ, Environment{ Name:  "oauth_client_secret_path", Value: "/var/secrets/of-client-secret/of-client-secret"})
	environ = append(environ, Environment{ Name:  "public_key_path", Value: "/var/secrets/public/key.pub"})
	environ = append(environ, Environment{ Name:  "private_key_path", Value: "/var/secrets/private/key"})
	environ = append(environ, Environment{ Name:  "client_secret", Value: "client-secret"})
	environ = append(environ, Environment{ Name:  "client_id", Value: "client-id"})
	environ = append(environ, Environment{ Name:  "oauth_provider_base_url", Value: ""})
	environ = append(environ, Environment{ Name:  "oauth_provider", Value: "github"})
	environ = append(environ, Environment{ Name:  "external_redirect_domain", Value: "https://auth.system.example.com"})
	environ = append(environ, Environment{ Name:  "cookie_root_domain", Value: ".system.example.com"})
	environ = append(environ, Environment{ Name:  "customers_url", Value: "https://example.com/customers_url"})
	environ = append(environ, Environment{ Name:  "write_debug", Value: "false"})

	return environ
}

func makeDeployVolumes(names []string) []DeploymentVolumes {
	var volumes []DeploymentVolumes

	for name := range names {
		volumes = append(volumes, DeploymentVolumes{
			SecretName: names[name],
			SecretInfo: map[string]string{"secretName": names[name]},
		})
	}



	return volumes
}