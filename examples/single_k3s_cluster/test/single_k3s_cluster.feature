Feature: Kubernetes Cluster
  In order to test Kubernetes clusters
  I should apply a blueprint
  And test the output

  Scenario: K3s Cluster
    Given I have a running blueprint
    Then the following resources should be running
      | name                      | type        |
      | cloud                     | network     |
      | server.k3s                | k8s_cluster |
    And a HTTP call to "http://localhost:18500/v1/agent/members" should result in status 200
    And a HTTP call to "http://localhost:18200" should result in status 200
    And a TCP connection to "localhost:8300" should open
