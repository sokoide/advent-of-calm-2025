package main

import "fmt"

func main() {
	arch := NewArchitecture("ecommerce-platform-architecture", "E-Commerce Order Processing Platform", "A complete architecture for an e-commerce order processing system.")

	addGeneralMetadata(arch)
	addGlobalControls(arch)
	addNodes(arch)
	addRelationships(arch)
	addFlows(arch)

	fmt.Println(arch.ToJSON())
}

func addGeneralMetadata(a *Architecture) {
	a.ADRs = []string{
		"docs/adr/0001-use-message-queue-for-async-processing.md",
		"docs/adr/0002-use-oauth2-for-api-authentication.md",
	}

	monitoring := NewMetadata().
		Add("grafana-dashboard", "https://grafana.example.com/d/ecommerce-overview").
		Add("kibana-logs", "https://kibana.example.com/app/discover#/ecommerce-*").
		Add("pagerduty-service", "https://pagerduty.example.com/services/ECOMMERCE").
		Add("statuspage", "https://status.example.com").
		Add("metrics-retention", "30 days").
		Add("log-retention", "90 days")

	a.Metadata = NewMetadata().
		Add("owner", "Architecture Team").
		Add("version", "1.0.0").
		Add("created", "2025-12-26").
		Add("description", "E-commerce order processing platform architecture demo.").
		Add("tags", []string{"ecommerce", "microservices", "orders"}).
		Add("monitoring", monitoring)
}

func addGlobalControls(a *Architecture) {
	a.Controls["security"] = NewControl("Data encryption and secure communication requirements",
		NewRequirement("https://internal-policy.example.com/security/encryption-at-rest", NewSecurityConfig("AES-256", "all-data-stores")),
		NewRequirementWithURL("https://internal-policy.example.com/security/tls-1-3-minimum", "https://configs.example.com/security/tls-config.yaml"),
	)
	a.Controls["performance"] = NewControl("System-wide performance and scalability requirements",
		NewRequirement("https://internal-policy.example.com/performance/response-time-sla", NewPerformanceConfig(200, 100)),
		NewRequirementWithURL("https://internal-policy.example.com/performance/availability-target", "https://configs.example.com/infra/ha-config.yaml"),
	)
	a.Controls["high-availability"] = NewControl("System-wide uptime and availability requirements",
		NewRequirement("https://internal-policy.example.com/resilience/availability-sla", NewAvailabilityConfig(99.9, 60)),
	)
}

func addFlows(a *Architecture) {
	flowMeta := NewMetadata().
		Add("business-impact", "Revenue loss").
		Add("degraded-behavior", "Queueing").
		Add("customer-communication", "Delayed msg").
		Add("sla", "99.9%")

	a.AddFlow(NewFlow("order-processing-flow", "Customer Order Processing", "End-to-end flow from customer placing an order to payment confirmation").
		WithMetadata(flowMeta).
		WithTransitions(
			NewTransition("customer-interacts-lb", 1, "Submit order", "source-to-destination"),
			NewTransition("lb-connects-gateway-1", 2, "Route to GW1", "source-to-destination"),
			NewTransition("gateway-1-connects-order", 3, "Route to Order Svc", "source-to-destination"),
			NewTransition("order-publishes-to-queue", 4, "Publish task", "source-to-destination"),
			NewTransition("payment-subscribes-to-queue", 5, "Process task", "source-to-destination"),
		))

	invFlowMeta := NewMetadata().
		Add("business-impact", "Overselling risk").
		Add("sla", "99.5%")

	a.AddFlow(NewFlow("inventory-check-flow", "Inventory Stock Check", "Admin checks and updates inventory stock levels").
		WithMetadata(invFlowMeta).
		WithTransitions(
			NewTransition("admin-interacts-lb", 1, "Request status", "source-to-destination"),
			NewTransition("lb-connects-gateway-2", 2, "Route to GW2", "source-to-destination"),
			NewTransition("gateway-2-connects-inventory", 3, "Route to inventory", "source-to-destination"),
			NewTransition("inventory-connects-db", 4, "Query stock", "source-to-destination"),
			NewTransition("inventory-connects-db", 5, "Return data", "destination-to-source"),
		))
}

func addNodes(a *Architecture) {
	a.AddNode(NewNode("customer", Actor, "Customer", "A user who browses and purchases products.").SetStandards("CC-1000", "marketing-team"))
	a.AddNode(NewNode("admin", Actor, "Admin", "A staff member who manages products and orders.").SetStandards("CC-1000", "ops-team"))
	a.AddNode(NewNode("ecommerce-system", System, "E-Commerce Platform", "Global system").SetStandards("CC-2000", "platform-team"))

	lbMeta := NewMetadata().
		Add("ha-enabled", true).
		Add("health-endpoint", "/status").
		Add("tier", "tier-1").
		Add("owner", "platform-team").
		Add("runbook", "https://runbooks.example.com/load-balancer")

	a.AddNode(NewNode("load-balancer", Service, "Load Balancer", "HA Entry").
		SetStandards("CC-2000", "platform-team").
		WithMetadata(lbMeta).
		WithInterfaces(
			NewInterface("lb-https", "HTTPS").AtPort(443).OnHost("api.shop.example.com"),
			NewInterface("lb-to-gateway", "HTTP"),
		))

	gwControls := map[string]Control{"performance": NewControl("GW performance",
		NewRequirementWithURL("https://rate-limiting", "https://limits.yaml"),
		NewRequirement("https://caching", NewMetadata().Add("default-ttl-seconds", 300)),
	)}

	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("api-gateway-%d", i)
		gwMeta := NewMetadata().
			Add("tier", "tier-1").
			Add("oncall-slack", "#oncall-platform").
			Add("dashboard", "https://grafana/gw").
			Add("log-query", fmt.Sprintf("service:api-gateway AND instance:%d", i))

		a.AddNode(NewNode(id, Service, fmt.Sprintf("Gateway %d", i), "Instance").
			SetStandards("CC-2000", "platform-team").
			WithMetadata(gwMeta).
			WithControl("performance", gwControls["performance"]).
			WithInterfaces(
				NewInterface(fmt.Sprintf("gateway-%d-http", i), "HTTP").AtPort(80),
				NewInterface(fmt.Sprintf("order-client-%d", i), "REST"),
				NewInterface(fmt.Sprintf("inventory-client-%d", i), "REST"),
			))
	}

	orderFailures := []Metadata{
		NewMetadata().Add("symptom", "HTTP 503").Add("likely-cause", "DB Pool Exhausted").Add("remediation", "Scale up"),
		NewMetadata().Add("symptom", "High latency").Add("likely-cause", "Payment Svc Down"),
	}

	a.AddNode(NewNode("order-service", Service, "Order Service", "Core Logic").
		SetStandards("CC-3000", "orders-team").
		WithMetadata(NewMetadata().Add("tier", "tier-1").Add("failure-modes", orderFailures)).
		WithControl("circuit-breaker", NewControl("Fault tolerance",
			NewRequirement("https://policy", NewCircuitBreakerConfig(50, 30, 10)),
		)).
		WithInterfaces(
			NewInterface("order-api", "REST").AtPort(8080),
			NewInterface("order-db-write-client", "JDBC"),
			NewInterface("order-db-read-client", "JDBC"),
			NewInterface("payment-publisher", "AMQP"),
			NewInterface("order-inventory-client", "REST"),
			NewInterface("order-health", "HTTP").WithPath("/health"),
		))

	a.AddNode(NewNode("inventory-service", Service, "Inventory Service", "Stock").
		SetStandards("CC-4000", "inventory-team").
		WithInterfaces(
			NewInterface("inventory-api", "REST").AtPort(8081),
			NewInterface("inventory-db-client", "JDBC"),
			NewInterface("inventory-health", "HTTP").WithPath("/health"),
		))

	a.AddNode(NewNode("payment-service", Service, "Payment Service", "Money").
		SetStandards("CC-5000", "payments-team").
		WithControl("compliance", NewControl("PCI-DSS",
			NewRequirementWithURL("https://pci-dss", "https://configs/pci.json"),
		)).
		WithInterfaces(
			NewInterface("payment-api", "REST").AtPort(8082),
			NewInterface("payment-consumer", "AMQP"),
			NewInterface("payment-health", "HTTP").WithPath("/health"),
		))

	a.AddNode(NewNode("message-broker", System, "RabbitMQ", "Broker").
		SetStandards("CC-2000", "platform-team").
		WithMetadata(NewMetadata().Add("adr", "docs/adr/0001-use-message-queue-for-async-processing.md")).
		WithInterfaces(NewInterface("amqp-port", "AMQP").AtPort(5672)))

	a.AddNode(NewNode("order-queue", Queue, "Order Queue", "Buffer").SetStandards("CC-3000", "orders-team"))

	a.AddNode(NewNode("order-database-cluster", System, "DB Cluster", "HA Data").
		SetStandards("CC-3000", "dba-team").
		WithControl("failover", NewControl("DR",
			NewRequirement("https://dr-targets", NewFailoverConfig(15, 5, true)),
		)))

	dbCommonMeta := NewMetadata().
		Add("tech-owner", "DBA Team").
		Add("backup-schedule", "daily").
		Add("data-classification", "PII")

	a.AddNode(NewNode("order-database-primary", Database, "Order DB Primary", "Write").
		SetStandards("CC-3000", "dba-team").
		WithMetadata(dbCommonMeta.Add("role", "primary")).
		WithInterfaces(NewInterface("order-sql-primary", "JDBC").AtPort(5432).WithDB("orders_v1")))

	a.AddNode(NewNode("order-database-replica", Database, "Order DB Replica", "Read").
		SetStandards("CC-3000", "dba-team").
		WithMetadata(dbCommonMeta.Add("role", "replica")).
		WithInterfaces(NewInterface("order-sql-replica", "JDBC").AtPort(5432).WithDB("orders_v1")))

	a.AddNode(NewNode("inventory-db", Database, "Inventory DB", "Stock data").
		SetStandards("CC-4000", "dba-team").
		WithMetadata(NewMetadata().Add("tech-owner", "DBA Team")).
		WithInterfaces(NewInterface("inventory-sql", "JDBC").AtPort(5432).WithDB("inventory_v1")))
}

func addRelationships(a *Architecture) {
	a.AddRelationship(Interacts("customer-interacts-lb", "Customer uses LB", "customer", "load-balancer").WithData("public", true))
	a.AddRelationship(Interacts("admin-interacts-lb", "Admin uses LB", "admin", "load-balancer").WithData("internal", true))

	a.AddRelationship(ConnectIntf("lb-connects-gateway-1", "LB to GW1", "load-balancer", "api-gateway-1", []string{"lb-to-gateway"}, []string{"gateway-1-http"}).WithData("internal", true))
	a.AddRelationship(ConnectIntf("lb-connects-gateway-2", "LB to GW2", "load-balancer", "api-gateway-2", []string{"lb-to-gateway"}, []string{"gateway-2-http"}).WithData("internal", true))

	a.AddRelationship(ConnectIntf("gateway-1-connects-order", "GW1 to Order", "api-gateway-1", "order-service", []string{"order-client-1"}, []string{"order-api"}).WithData("internal", true))
	a.AddRelationship(ConnectIntf("gateway-2-connects-order", "GW2 to Order", "api-gateway-2", "order-service", []string{"order-client-2"}, []string{"order-api"}).WithData("internal", true))
	a.AddRelationship(ConnectIntf("gateway-1-connects-inventory", "GW1 to Inventory", "api-gateway-1", "inventory-service", []string{"inventory-client-1"}, []string{"inventory-api"}).WithData("internal", true))
	a.AddRelationship(ConnectIntf("gateway-2-connects-inventory", "GW2 to Inventory", "api-gateway-2", "inventory-service", []string{"inventory-client-2"}, []string{"inventory-api"}).WithData("internal", true))

	a.AddRelationship(ConnectIntf("order-connects-primary-db", "Order Write", "order-service", "order-database-primary", []string{"order-db-write-client"}, []string{"order-sql-primary"}).WithData("confidential", true).WithMetadata(NewMetadata().Add("monitoring", true)))
	a.AddRelationship(ConnectIntf("order-publishes-to-queue", "Pub", "order-service", "order-queue", []string{"payment-publisher"}, nil).WithData("internal", true).WithProtocol("AMQP"))
	a.AddRelationship(ConnectIntf("payment-subscribes-to-queue", "Sub", "order-queue", "payment-service", nil, []string{"payment-consumer"}).WithData("internal", true).WithProtocol("AMQP"))
	a.AddRelationship(ConnectIntf("order-connects-replica-db", "Order Read", "order-service", "order-database-replica", []string{"order-db-read-client"}, []string{"order-sql-replica"}).WithData("confidential", true).WithMetadata(NewMetadata().Add("monitoring", true)))
	a.AddRelationship(ConnectIntf("order-connects-inventory", "Order Check Stock", "order-service", "inventory-service", []string{"order-inventory-client"}, []string{"inventory-api"}).WithData("internal", true).WithMetadata(NewMetadata().Add("monitoring", true).Add("circuit-breaker", true)))
	a.AddRelationship(ConnectIntf("inventory-connects-db", "Inventory Stock Persistence", "inventory-service", "inventory-db", []string{"inventory-db-client"}, []string{"inventory-sql"}).WithData("internal", true).WithMetadata(NewMetadata().Add("monitoring", true)))

	a.AddRelationship(ComposedOf("broker-composition", "Broker has queue", "message-broker", []string{"order-queue"}).WithData("internal", false))
	a.AddRelationship(ComposedOf("order-db-composition", "Cluster has DBs", "order-database-cluster", []string{"order-database-primary", "order-database-replica"}).WithData("confidential", true))
	a.AddRelationship(ComposedOf("ecommerce-system-composition", "Platform composed", "ecommerce-system", []string{"load-balancer", "api-gateway-1", "api-gateway-2", "order-service", "inventory-service", "payment-service", "order-database-cluster", "inventory-db", "message-broker"}).WithData("internal", false))
}
