Feature: Docker Container
  In order to test Shipyard creates containers correctly
  I should apply a blueprint which defines a simple container setup
  and test the resources are created correctly

Scenario: Single Container from Local Blueprint
  Given I have a running blueprint
  Then the following resources should be running
    | name                      | type                |
    | onprem                    | network             |
    | consul                    | container           |
    | envoy                     | sidecar             |
    | consul-container-http     | container_ingress   |
  And the info "{.NetworkSettings.Ports['8501/tcp']}" for the running "container" called "consul" should exist"
  And the info "{.NetworkSettings.Ports['8500/tcp'][0].HostPort}" for the running "container" called "consul" should equal "8500"
  And the info "{.NetworkSettings.Ports['8500/tcp'][0].HostPort}" for the running "container" called "consul" should contain "85"
  And a HTTP call to "http://consul.container.shipyard.run:8500/v1/status/leader" should result in status 200
  And a HTTP call to "http://consul-http.ingress.shipyard.run:28500/v1/status/leader" should result in status 200

Scenario: Single Container from Local Blueprint with multiple runs
  Given the environment variable "CONSUL_VERSION" has a value "<consul>"
  And the environment variable "ENVOY_VERSION" has a value "<envoy>"
  And I have a running blueprint
  Then the following resources should be running
    | name                      | type                |
    | onprem                    | network             |
    | consul                    | container           |
    | envoy                     | sidecar             |
    | consul-container-http     | container_ingress   |
  And a HTTP call to "http://consul-http.ingress.shipyard.run:8500/v1/status/leader" should result in status 200
  And the response body should contain "10.6.0.200"
  When I run the command "curl -k http://1.server.dc1.consul.container.shipyard.run:8500/v1/status/leader"
  Then I expect the exit code to be 0
  Examples:
    | consul            | envoy    |
    | 1.11.1            | 1.18.4   |
    | 1.10.6            | 1.18.4   |
