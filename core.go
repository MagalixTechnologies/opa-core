package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
)

// ParseWithRule constructs OPA policy from string
func ParseWithRule(content, ruleName string) (Policy, error) {
	// validate module
	module, err := ast.ParseModule("", content)
	if err != nil {
		return Policy{}, err
	}

	if module == nil {
		return Policy{}, fmt.Errorf("Failed to parse module: empty content")
	}

	var valid bool
	for _, rule := range module.Rules {
		if rule.Head.Name == ast.Var(ruleName) {
			valid = true
			break
		}
	}

	if !valid {
		return Policy{}, fmt.Errorf("rule `%s` is not found", ruleName)
	}

	policy := Policy{
		module: module,
		pkg:    strings.Split(module.Package.String(), "package ")[1],
	}

	err = policy.Eval("{}", ruleName)
	var opaErr OPAError
	if err != nil && !errors.As(err, &opaErr) {
		return Policy{}, err
	}

	valiationSpec1 := `{"kind":"Deployment","spec":{"replicas":1,"selector":{"matchLabels":{"app":"wordpress-wordpress","release":"wordpress"}},"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxSurge":"25%","maxUnavailable":"25%"}},"template":{"spec":{"volumes":[{"name":"wordpress-data","persistentVolumeClaim":{"claimName":"wordpress-wordpress"}}],"dnsPolicy":"ClusterFirst","containers":[{"env":[{"name":"ALLOW_EMPTY_PASSWORD","value":"yes"},{"name":"MARIADB_HOST","value":"wordpress-mariadb"},{"name":"MARIADB_PORT_NUMBER","value":"3306"},{"name":"WORDPRESS_DATABASE_NAME","value":"bitnami_wordpress"},{"name":"WORDPRESS_DATABASE_USER","value":"bn_wordpress"},{"name":"WORDPRESS_DATABASE_PASSWORD","valueFrom":{"secretKeyRef":{"key":"mariadb-password","name":"wordpress-mariadb"}}},{"name":"WORDPRESS_USERNAME","value":"user"},{"name":"WORDPRESS_PASSWORD","valueFrom":{"secretKeyRef":{"key":"wordpress-password","name":"wordpress-wordpress"}}},{"name":"WORDPRESS_EMAIL","value":"user@example.com"},{"name":"WORDPRESS_FIRST_NAME","value":"FirstName"},{"name":"WORDPRESS_LAST_NAME","value":"LastName"},{"name":"WORDPRESS_HTACCESS_OVERRIDE_NONE","value":"no"},{"name":"WORDPRESS_BLOG_NAME","value":"User's Blog!"},{"name":"WORDPRESS_SKIP_INSTALL","value":"no"},{"name":"WORDPRESS_TABLE_PREFIX","value":"wp_"}],"name":"wordpress","image":"docker.io/bitnami/wordpress:5.2.1-debian-9-r1","ports":[{"name":"http","protocol":"TCP","containerPort":80},{"name":"https","protocol":"TCP","containerPort":443}],"resources":{"limits":{"cpu":"1","memory":"450Mi"},"requests":{"cpu":"50m","memory":"300Mi"}},"volumeMounts":[{"name":"wordpress-data","subPath":"apache","mountPath":"/bitnami/apache"},{"name":"wordpress-data","subPath":"wordpress","mountPath":"/bitnami/wordpress"},{"name":"wordpress-data","subPath":"php","mountPath":"/bitnami/php"}],"livenessProbe":{"httpGet":{"path":"/wp-login.php","port":"http","scheme":"HTTP"},"periodSeconds":10,"timeoutSeconds":5,"failureThreshold":6,"successThreshold":1,"initialDelaySeconds":120},"readinessProbe":{"httpGet":{"path":"/wp-login.php","port":"http","scheme":"HTTP"},"periodSeconds":10,"timeoutSeconds":5,"failureThreshold":6,"successThreshold":1,"initialDelaySeconds":30},"imagePullPolicy":"IfNotPresent","terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File"}],"hostAliases":[{"ip":"127.0.0.1","hostnames":["status.localhost"]}],"restartPolicy":"Always","schedulerName":"default-scheduler","securityContext":{},"terminationGracePeriodSeconds":30},"metadata":{"labels":{"app":"wordpress-wordpress","chart":"wordpress-5.9.6","release":"wordpress"},"creationTimestamp":null}},"revisionHistoryLimit":10,"progressDeadlineSeconds":600},"status":{"replicas":2,"conditions":[{"type":"Available","reason":"MinimumReplicasUnavailable","status":"False","message":"Deployment does not have minimum availability.","lastUpdateTime":"2019-09-19T21:53:15Z","lastTransitionTime":"2019-09-19T21:53:15Z"},{"type":"Progressing","reason":"ProgressDeadlineExceeded","status":"False","message":"ReplicaSet \"wordpress-wordpress-6ffddd49ff\" has timed out progressing.","lastUpdateTime":"2020-11-12T15:22:39Z","lastTransitionTime":"2020-11-12T15:22:39Z"}],"updatedReplicas":1,"observedGeneration":8,"unavailableReplicas":2},"metadata":{"uid":"34f089bb-8ebc-11e9-a8f5-06e4f842c366","name":"wordpress-wordpress","labels":{"app":"wordpress-wordpress","chart":"wordpress-5.9.6","release":"wordpress","heritage":"Tiller"},"selfLink":"/apis/apps/v1/namespaces/default/deployments/wordpress-wordpress","namespace":"default","generation":8,"annotations":{"deployment.kubernetes.io/revision":"8"},"resourceVersion":"63694454","creationTimestamp":"2019-06-14T15:51:00Z"},"apiVersion":"apps/v1"}`
	jsonMap := make(map[string]interface{})
	err = json.Unmarshal([]byte(valiationSpec1), &jsonMap)
	if err != nil {
		return Policy{}, err
	}
	err = policy.Eval(jsonMap, ruleName)
	if err != nil && !errors.As(err, &opaErr) {
		return Policy{}, err
	}

	validationSpec2 := `{"kind":"Deployment","spec":{"replicas":1,"selector":{"matchLabels":{"k8s-app":"metrics-server","version":"v0.3.6"}},"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxSurge":"25%","maxUnavailable":"25%"}},"template":{"spec":{"volumes":[{"name":"metrics-server-config-volume","configMap":{"name":"metrics-server-config","defaultMode":420}}],"dnsPolicy":"ClusterFirst","containers":[{"name":"metrics-server","image":"k8s.gcr.io/metrics-server-amd64:v0.3.6","ports":[{"name":"https","protocol":"TCP","containerPort":443}],"command":["/metrics-server","--metric-resolution=30s","--kubelet-port=10255","--deprecated-kubelet-completely-insecure=true","--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP"],"resources":{"limits":{"cpu":"45m","memory":"75Mi"},"requests":{"cpu":"45m","memory":"75Mi"}},"imagePullPolicy":"IfNotPresent","terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File"},{"env":[{"name":"MY_POD_NAME","valueFrom":{"fieldRef":{"fieldPath":"metadata.name","apiVersion":"v1"}}},{"name":"MY_POD_NAMESPACE","valueFrom":{"fieldRef":{"fieldPath":"metadata.namespace","apiVersion":"v1"}}}],"name":"metrics-server-nanny","image":"gke.gcr.io/addon-resizer:1.8.8-gke.1","command":["/pod_nanny","--config-dir=/etc/config","--cpu=40m","--extra-cpu=0.5m","--memory=35Mi","--extra-memory=4Mi","--threshold=5","--deployment=metrics-server-v0.3.6","--container=metrics-server","--poll-period=300000","--estimator=exponential","--scale-down-delay=24h","--minClusterSize=5"],"resources":{"limits":{"cpu":"100m","memory":"300Mi"},"requests":{"cpu":"5m","memory":"50Mi"}},"volumeMounts":[{"name":"metrics-server-config-volume","mountPath":"/etc/config"}],"imagePullPolicy":"IfNotPresent","terminationMessagePath":"/dev/termination-log","terminationMessagePolicy":"File"}],"tolerations":[{"key":"CriticalAddonsOnly","operator":"Exists"},{"key":"components.gke.io/gke-managed-components","operator":"Exists"}],"nodeSelector":{"kubernetes.io/os":"linux"},"restartPolicy":"Always","schedulerName":"default-scheduler","serviceAccount":"metrics-server","securityContext":{},"priorityClassName":"system-cluster-critical","serviceAccountName":"metrics-server","terminationGracePeriodSeconds":30},"metadata":{"name":"metrics-server","labels":{"k8s-app":"metrics-server","version":"v0.3.6"},"annotations":{"seccomp.security.alpha.kubernetes.io/pod":"docker/default"},"creationTimestamp":null}},"revisionHistoryLimit":10,"progressDeadlineSeconds":600},"status":{"replicas":1,"conditions":[{"type":"Available","reason":"MinimumReplicasAvailable","status":"True","message":"Deployment has minimum availability.","lastUpdateTime":"2020-09-22T09:45:25Z","lastTransitionTime":"2020-09-22T09:45:25Z"},{"type":"Progressing","reason":"NewReplicaSetAvailable","status":"True","message":"ReplicaSet \"metrics-server-v0.3.6-75fbc64cff\" has successfully progressed.","lastUpdateTime":"2020-10-05T09:18:11Z","lastTransitionTime":"2020-09-22T09:25:12Z"}],"readyReplicas":1,"updatedReplicas":1,"availableReplicas":1,"observedGeneration":4},"metadata":{"uid":"a0e89fdf-e888-4fd3-87d9-f6ece139f738","name":"metrics-server-v0.3.6","labels":{"k8s-app":"metrics-server","version":"v0.3.6","kubernetes.io/cluster-service":"true","addonmanager.kubernetes.io/mode":"Reconcile"},"selfLink":"/apis/apps/v1/namespaces/kube-system/deployments/metrics-server-v0.3.6","namespace":"kube-system","generation":4,"annotations":{"deployment.kubernetes.io/revision":"4","kubectl.kubernetes.io/last-applied-configuration":"{\"apiVersion\":\"apps/v1\",\"kind\":\"Deployment\",\"metadata\":{\"annotations\":{},\"labels\":{\"addonmanager.kubernetes.io/mode\":\"Reconcile\",\"k8s-app\":\"metrics-server\",\"kubernetes.io/cluster-service\":\"true\",\"version\":\"v0.3.6\"},\"name\":\"metrics-server-v0.3.6\",\"namespace\":\"kube-system\"},\"spec\":{\"selector\":{\"matchLabels\":{\"k8s-app\":\"metrics-server\",\"version\":\"v0.3.6\"}},\"template\":{\"metadata\":{\"annotations\":{\"seccomp.security.alpha.kubernetes.io/pod\":\"docker/default\"},\"labels\":{\"k8s-app\":\"metrics-server\",\"version\":\"v0.3.6\"},\"name\":\"metrics-server\"},\"spec\":{\"containers\":[{\"command\":[\"/metrics-server\",\"--metric-resolution=30s\",\"--kubelet-port=10255\",\"--deprecated-kubelet-completely-insecure=true\",\"--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP\"],\"image\":\"k8s.gcr.io/metrics-server-amd64:v0.3.6\",\"name\":\"metrics-server\",\"ports\":[{\"containerPort\":443,\"name\":\"https\",\"protocol\":\"TCP\"}]},{\"command\":[\"/pod_nanny\",\"--config-dir=/etc/config\",\"--cpu=40m\",\"--extra-cpu=0.5m\",\"--memory=35Mi\",\"--extra-memory=4Mi\",\"--threshold=5\",\"--deployment=metrics-server-v0.3.6\",\"--container=metrics-server\",\"--poll-period=300000\",\"--estimator=exponential\",\"--scale-down-delay=24h\",\"--minClusterSize=5\"],\"env\":[{\"name\":\"MY_POD_NAME\",\"valueFrom\":{\"fieldRef\":{\"fieldPath\":\"metadata.name\"}}},{\"name\":\"MY_POD_NAMESPACE\",\"valueFrom\":{\"fieldRef\":{\"fieldPath\":\"metadata.namespace\"}}}],\"image\":\"gke.gcr.io/addon-resizer:1.8.8-gke.1\",\"name\":\"metrics-server-nanny\",\"resources\":{\"limits\":{\"cpu\":\"100m\",\"memory\":\"300Mi\"},\"requests\":{\"cpu\":\"5m\",\"memory\":\"50Mi\"}},\"volumeMounts\":[{\"mountPath\":\"/etc/config\",\"name\":\"metrics-server-config-volume\"}]}],\"nodeSelector\":{\"kubernetes.io/os\":\"linux\"},\"priorityClassName\":\"system-cluster-critical\",\"serviceAccountName\":\"metrics-server\",\"tolerations\":[{\"key\":\"CriticalAddonsOnly\",\"operator\":\"Exists\"},{\"key\":\"components.gke.io/gke-managed-components\",\"operator\":\"Exists\"}],\"volumes\":[{\"configMap\":{\"name\":\"metrics-server-config\"},\"name\":\"metrics-server-config-volume\"}]}}}}\n"},"resourceVersion":"462934578","creationTimestamp":"2020-09-22T09:25:12Z"},"apiVersion":"apps/v1"}`
	jsonMap = make(map[string]interface{})
	err = json.Unmarshal([]byte(validationSpec2), &jsonMap)
	if err != nil {
		return Policy{}, err
	}

	err = policy.Eval(jsonMap, ruleName)
	if err != nil && !errors.As(err, &opaErr) {
		return Policy{}, err
	}

	return policy, nil
}

// Parse constructs OPA policy from string with violations as rule name
func Parse(content string) (Policy, error) {
	return ParseWithRule(content, "violations")
}

// Eval validates data against given policy
// returns error if there're any violations found
func (p Policy) Eval(data interface{}, query string) error {
	rego := rego.New(
		rego.Query(fmt.Sprintf("data.%s.%s", p.pkg, query)),
		rego.ParsedModule(p.module),
		rego.Input(data),
	)

	// Run evaluation.
	rs, err := rego.Eval(context.Background())
	if err != nil {
		return err
	}
	for _, r := range rs {
		for _, expr := range r.Expressions {
			switch reflect.TypeOf(expr.Value).Kind() {
			// FIXME: support more formats
			case reflect.Slice:
				s := expr.Value.([]interface{})
				// FIXME: return multiple violations if found
				for i := 0; i < len(s); i++ {
					err := NoValidError{
						Details: s[i],
					}
					return err
				}
			case reflect.Map:
				s := expr.Value.(map[string]interface{})
				err := NoValidError{
					Details: s,
				}
				return err
			}
		}
	}

	return nil
}
