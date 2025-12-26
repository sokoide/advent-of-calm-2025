package main

import "fmt"

func main() {
	arch := NewArchitecture(
		"ecommerce-platform-architecture",
		"E-Commerce Order Processing Platform",
		"A complete architecture for an e-commerce order processing system.",
	)

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
	a.Controls["security"] = NewControl(
		"Data encryption and secure communication requirements",
		NewRequirement(
			"https://internal-policy.example.com/security/encryption-at-rest",
			NewSecurityConfig("AES-256", "all-data-stores"),
		),
		NewRequirementWithURL(
			"https://internal-policy.example.com/security/tls-1-3-minimum",
			"https://configs.example.com/security/tls-config.yaml",
		),
	)
	a.Controls["performance"] = NewControl(
		"System-wide performance and scalability requirements",
		NewRequirement(
			"https://internal-policy.example.com/performance/response-time-sla",
			NewPerformanceConfig(200, 100),
		),
		NewRequirementWithURL(
			"https://internal-policy.example.com/performance/availability-target",
			"https://configs.example.com/infra/ha-config.yaml",
		),
	)
	a.Controls["high-availability"] = NewControl(
		"System-wide uptime and availability requirements",
		NewRequirement(
			"https://internal-policy.example.com/resilience/availability-sla",
			NewAvailabilityConfig(99.9, 60),
		),
	)
}

func addFlows(a *Architecture) {
	a.AddFlow(NewFlow(
		"order-processing-flow",
		"Customer Order Processing",
		"End-to-end flow from customer placing an order to payment confirmation",
	).WithMetadata(NewMetadata().
		Add("business-impact", "Customers cannot complete purchases - direct revenue loss").
		Add("degraded-behavior", "Orders queue in message broker; processed when service recovers").
		Add("customer-communication", "Display 'Order processing delayed' message").
		Add("sla", "99.9% availability, 30s p99 latency")).
		WithTransitions(
			NewTransition(
				"customer-interacts-lb",
				1,
				"Customer submits order via Load Balancer",
				"source-to-destination",
			),
			NewTransition("lb-connects-gateway-1", 2, "LB routes to Gateway 1", "source-to-destination"),
			NewTransition("gateway-1-connects-order", 3, "Gateway 1 routes to Order Service", "source-to-destination"),
			NewTransition(
				"order-publishes-to-queue",
				4,
				"Order Service publishes payment task",
				"source-to-destination",
			),
			NewTransition(
				"payment-subscribes-to-queue",
				5,
				"Payment Service processes task from queue",
				"source-to-destination",
			),
		),
	)

	a.AddFlow(NewFlow(
		"inventory-check-flow",
		"Inventory Stock Check",
		"Admin checks and updates inventory stock levels",
	).WithMetadata(NewMetadata().
		Add("business-impact", "Stock levels may be inaccurate - risk of overselling").
		Add("degraded-behavior", "Fall back to cached inventory; flag orders for manual review").
		Add("customer-communication", "Display 'Stock availability being confirmed'").
		Add("sla", "99.5% availability, 500ms p99 latency")).
		WithTransitions(
			NewTransition("admin-interacts-lb", 1, "Admin requests inventory status via LB", "source-to-destination"),
			NewTransition("lb-connects-gateway-2", 2, "LB routes to Gateway 2", "source-to-destination"),
			NewTransition(
				"gateway-2-connects-inventory",
				3,
				"Gateway 2 routes to inventory service",
				"source-to-destination",
			),
			NewTransition("inventory-connects-db", 4, "Query current stock levels", "source-to-destination"),
			NewTransition("inventory-connects-db", 5, "Return stock data", "destination-to-source"),
		),
	)
}

func addNodes(a *Architecture) {
	// --- Actors ---
	a.AddNode(NewNode("customer", Actor, "Customer", "A user who browses and purchases products.").
		SetStandards("CC-1000", "marketing-team"))
	a.AddNode(NewNode("admin", Actor, "Admin", "A staff member who manages products and orders.").
		SetStandards("CC-1000", "ops-team"))
	a.AddNode(
		NewNode(
			"ecommerce-system",
			System,
			"E-Commerce Platform",
			"The overall e-commerce system containing microservices.",
		).
			SetStandards("CC-2000", "platform-team"),
	)

	// --- Infrastructure ---
	a.AddNode(
		NewNode(
			"load-balancer",
			Service,
			"Load Balancer",
			"High-availability entry point that distributes traffic to API Gateways.",
		).
			SetStandards("CC-2000", "platform-team").
			WithMetadata(NewMetadata().
				Add("tech-owner", "Network Team").
				Add("owner", "platform-team").
				Add("oncall-slack", "#oncall-platform").
				Add("health-endpoint", "/status").
				Add("runbook", "https://runbooks.example.com/load-balancer").
				Add("tier", "tier-1").
				Add("dependencies", []string{"api-gateway-1", "api-gateway-2"}).
				Add("dashboard", "https://grafana.example.com/d/lb-metrics").
				Add("log-query", "service:load-balancer AND error").
				Add("alerts", []string{"LB-HighLatency", "LB-TargetGroupUnhealthy"}).
				Add("deployment-type", "managed-service").
				Add("business-criticality", "high").
				Add("ha-enabled", true)).
			WithInterfaces(
				NewInterface(
					"lb-https",
					"HTTPS",
				).WithName("Public HTTPS Interface").
					AtPort(443).
					OnHost("api.shop.example.com"),
				NewInterface("lb-to-gateway", "HTTP").WithDesc("Outbound to API Gateways"),
			),
	)

	gwControls := map[string]Control{"performance": NewControl(
		"API Gateway rate limiting and caching requirements",
		NewRequirementWithURL(
			"https://internal-policy.example.com/performance/rate-limiting",
			"https://configs.example.com/gateway/rate-limits.yaml",
		),
		NewRequirement(
			"https://internal-policy.example.com/performance/caching-policy",
			NewMetadata().Add("default-ttl-seconds", 300).Add("cache-control", "private"),
		),
	)}

	// API Gateway 1
	a.AddNode(NewNode("api-gateway-1", Service, "API Gateway Instance 1", "Primary API Gateway instance.").
		SetStandards("CC-2000", "platform-team").
		WithMetadata(NewMetadata().
			Add("tech-owner", "Edge Team").
			Add("owner", "platform-team").
			Add("oncall-slack", "#oncall-platform").
			Add("health-endpoint", "/health").
			Add("runbook", "https://runbooks.example.com/api-gateway").
			Add("tier", "tier-1").
			Add("dependencies", []string{"order-service", "inventory-service"}).
			Add("dashboard", "https://grafana.example.com/d/gateway-overview").
			Add("log-query", "service:api-gateway AND instance:1").
			Add("alerts", []string{"Gateway-5xx-Rate", "Gateway-HighLatency"}).
			Add("repository", "https://github.com/example/api-gateway").
			Add("deployment-type", "container").
			Add("business-criticality", "high")).
		WithControl("performance", gwControls["performance"]).
		WithInterfaces(
			NewInterface("gateway-1-http", "HTTP").WithName("HTTP Interface").AtPort(80),
			NewInterface("order-client-1", "REST"),
			NewInterface("inventory-client-1", "REST"),
			NewInterface("gateway-1-health", "HTTP").WithName("Health Check").WithPath("/health"),
		))

	// API Gateway 2
	a.AddNode(
		NewNode(
			"api-gateway-2",
			Service,
			"API Gateway Instance 2",
			"Secondary API Gateway instance for high availability.",
		).
			SetStandards("CC-2000", "platform-team").
			WithMetadata(NewMetadata().
				Add("tech-owner", "Edge Team").
				Add("owner", "platform-team").
				Add("oncall-slack", "#oncall-platform").
				Add("health-endpoint", "/health").
				Add("runbook", "https://runbooks.example.com/api-gateway").
				Add("tier", "tier-1").
				Add("dependencies", []string{"order-service", "inventory-service"}).
				Add("dashboard", "https://grafana.example.com/d/gateway-overview").
				Add("log-query", "service:api-gateway AND instance:2").
				Add("alerts", []string{"Gateway-5xx-Rate", "Gateway-HighLatency"}).
				Add("repository", "https://github.com/example/api-gateway").
				Add("deployment-type", "container").
				Add("business-criticality", "high")).
			WithControl("performance", gwControls["performance"]).
			WithInterfaces(
				NewInterface("gateway-2-http", "HTTP").WithName("HTTP Interface").AtPort(80),
				NewInterface("order-client-2", "REST"),
				NewInterface("inventory-client-2", "REST"),
				NewInterface("gateway-2-health", "HTTP").WithName("Health Check").WithPath("/health"),
			),
	)

	// --- Core Services ---
	orderFailures := []Metadata{
		NewMetadata().
			Add("check", "Check connection pool metrics in Grafana dashboard").
			Add("escalation", "If persists > 5min, page DBA team").
			Add("likely-cause", "Database connection pool exhausted").
			Add("remediation", "Scale up service replicas or increase pool size").
			Add("symptom", "HTTP 503 errors"),
		NewMetadata().
			Add("check", "Check payment-service health and circuit breaker status").
			Add("escalation", "Contact payments-team if circuit breaker not triggering").
			Add("likely-cause", "Payment service degradation").
			Add("remediation", "Circuit breaker should open automatically; check fallback queue").
			Add("symptom", "High latency (>2s p99)"),
		NewMetadata().
			Add("check", "Verify inventory-service cache TTL and database replication lag").
			Add("escalation", "Contact platform-team for cache issues").
			Add("likely-cause", "Inventory service returning stale data").
			Add("remediation", "Clear inventory cache; check replica sync status").
			Add("symptom", "Order validation failures"),
	}

	a.AddNode(NewNode("order-service", Service, "Order Service", "Handles order creation and lifecycle management.").
		SetStandards("CC-3000", "orders-team").
		WithMetadata(NewMetadata().
			Add("alerts", []string{"OrderCreationFailureRate", "OrderDBConectionExhausted"}).
			Add("business-criticality", "high").
			Add("dashboard", "https://grafana.example.com/d/order-service-metrics").
			Add("dependencies", []string{"order-database-cluster", "inventory-service", "message-broker"}).
			Add("deployment-type", "container").
			Add("failure-modes", orderFailures).
			Add("health-endpoint", "/actuator/health").
			Add("log-query", "app:order-service AND level:ERROR").
			Add("oncall-slack", "#oncall-orders").
			Add("owner", "orders-team").
			Add("repository", "https://github.com/example/order-service").
			Add("runbook", "https://runbooks.example.com/order-service").
			Add("sla-tier", "tier-1").
			Add("tech-owner", "Order Team").
			Add("tier", "tier-1")).
		WithControl("circuit-breaker", NewControl(
			"Fault tolerance for downstream service calls",
			NewRequirement(
				"https://internal-policy.example.com/resilience/circuit-breaker-policy",
				NewCircuitBreakerConfig(50, 30, 10),
			),
		)).
		WithInterfaces(
			NewInterface("order-api", "REST").WithName("Order API").AtPort(8080),
			NewInterface("order-db-write-client", "JDBC").
				WithName("Order DB Write Client").
				WithDesc("Writes to the primary database."),
			NewInterface("order-db-read-client", "JDBC").
				WithName("Order DB Read Client").
				WithDesc("Reads from the replica database."),
			NewInterface("payment-publisher", "AMQP").
				WithDesc("Publishes order messages for payment processing."),
			NewInterface("order-inventory-client", "REST").
				WithDesc("Outbound connection to check inventory"),
			NewInterface("order-health", "HTTP").WithName("Health Check").WithPath("/health"),
		))

	invFailures := []Metadata{
		NewMetadata().
			Add("check", "Check DB lock metrics and slow query log").
			Add("escalation", "Contact DBA team for lock contention").
			Add("likely-cause", "Deadlock on stock updates").
			Add("remediation", "Review transaction isolation level or retry logic").
			Add("symptom", "Inventory sync failures"),
		NewMetadata().
			Add("check", "Verify Redis/Memcached availability and evictions").
			Add("escalation", "Contact platform-team for cache infrastructure").
			Add("likely-cause", "Cache invalidation failure").
			Add("remediation", "Flush cache for affected products").
			Add("symptom", "Stale stock levels"),
	}

	a.AddNode(NewNode("inventory-service", Service, "Inventory Service", "Manages product stock levels.").
		SetStandards("CC-4000", "inventory-team").
		WithMetadata(NewMetadata().
			Add("alerts", []string{"InventoryCacheInconsistency", "StockUpdateFailure"}).
			Add("business-criticality", "high").
			Add("dashboard", "https://grafana.example.com/d/inventory-metrics").
			Add("dependencies", []string{"inventory-db"}).
			Add("deployment-type", "container").
			Add("failure-modes", invFailures).
			Add("health-endpoint", "/health").
			Add("log-query", "app:inventory-service").
			Add("oncall-slack", "#oncall-inventory").
			Add("owner", "inventory-team").
			Add("repository", "https://github.com/example/inventory-service").
			Add("runbook", "https://runbooks.example.com/inventory-service").
			Add("tech-owner", "Warehouse Team").
			Add("tier", "tier-2")).
		WithInterfaces(
			NewInterface("inventory-api", "REST").WithName("Inventory API").AtPort(8081),
			NewInterface("inventory-db-client", "JDBC"),
			NewInterface("inventory-health", "HTTP").WithName("Health Check").WithPath("/health"),
		))

	payFailures := []Metadata{
		NewMetadata().
			Add("check", "Verify external gateway status page").
			Add("escalation", "Escalate to provider support").
			Add("likely-cause", "External payment gateway latency").
			Add("remediation", "Enable aggressive retry for idempotent calls").
			Add("symptom", "Payment processing timeouts"),
		NewMetadata().
			Add("check", "Review access logs for unusual patterns").
			Add("escalation", "Contact security-team").
			Add("likely-cause", "API Key leaked or compromised").
			Add("remediation", "Rotate API keys immediately").
			Add("symptom", "Unauthorized transaction spikes"),
	}

	a.AddNode(NewNode("payment-service", Service, "Payment Service", "Integrates with external payment providers.").
		SetStandards("CC-5000", "payments-team").
		WithMetadata(NewMetadata().
			Add("alerts", []string{"PaymentGatewayTimeout", "PCIViolationAttempt"}).
			Add("business-criticality", "high").
			Add("dashboard", "https://grafana.example.com/d/payment-metrics").
			Add("dependencies", []string{"external-payment-provider"}).
			Add("deployment-type", "serverless").
			Add("failure-modes", payFailures).
			Add("health-endpoint", "/health").
			Add("log-query", "app:payment-service").
			Add("oncall-slack", "#oncall-payments").
			Add("owner", "payments-team").
			Add("repository", "https://github.com/example/payment-service").
			Add("runbook", "https://runbooks.example.com/payment-service").
			Add("tech-owner", "Payment Team").
			Add("tier", "tier-1")).
		WithControl("compliance", NewControl(
			"PCI-DSS compliance for payment processing",
			NewRequirementWithURL(
				"https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0",
				"https://configs.example.com/compliance/pci-dss-config.json",
			),
		)).
		WithInterfaces(
			NewInterface("payment-api", "REST").WithName("Payment Processing API").AtPort(8082),
			NewInterface("payment-consumer", "AMQP").
				WithDesc("Consumes order messages for payment processing."),
			NewInterface("payment-health", "HTTP").WithName("Health Check").WithPath("/health"),
		))

	// --- Messaging ---
	a.AddNode(
		NewNode(
			"message-broker",
			System,
			"Message Broker (RabbitMQ)",
			"Central messaging system for failure isolation and async processing.",
		).
			SetStandards("CC-2000", "platform-team").
			WithMetadata(NewMetadata().
				Add("adr", "docs/adr/0001-use-message-queue-for-async-processing.md").
				Add("alerts", []string{"RabbitMQQueueSizeHigh", "RabbitMQConsumerDown"}).
				Add("dashboard", "https://grafana.example.com/d/rabbitmq-overview").
				Add("deployment-type", "managed-service").
				Add("health-endpoint", "/health").
				Add("log-query", "service:rabbitmq").
				Add("oncall-slack", "#oncall-platform").
				Add("owner", "platform-team").
				Add("runbook", "https://runbooks.example.com/rabbitmq").
				Add("tech-owner", "Platform Team").
				Add("tier", "tier-1")).
			WithInterfaces(NewInterface("amqp-port", "AMQP").AtPort(5672)),
	)

	a.AddNode(NewNode("order-queue", Queue, "Order Payment Queue", "Buffer for orders awaiting payment processing.").
		SetStandards("CC-3000", "orders-team"))

	// --- Databases ---
	a.AddNode(
		NewNode(
			"order-database-cluster",
			System,
			"Order Database Cluster",
			"High-availability database cluster for order data.",
		).
			SetStandards("CC-3000", "dba-team").
			WithControl("failover", NewControl(
				"Disaster recovery and failover targets",
				NewRequirement(
					"https://internal-policy.example.com/resilience/disaster-recovery-targets",
					NewFailoverConfig(15, 5, true),
				),
			)),
	)

	dbMeta := func(role string) Metadata {
		return NewMetadata().
			Add("backup-schedule", "daily at 02:00 UTC").
			Add("data-classification", "PII").
			Add("dba-contact", "dba-team@example.com").
			Add("deployment-type", "managed-service").
			Add("replication-mode", "async").
			Add("restore-time", "60 minutes").
			Add("role", role).
			Add("tech-owner", "DBA Team")
	}

	a.AddNode(
		NewNode("order-database-primary", Database, "Order Database (Primary)", "Main writable database for orders.").
			SetStandards("CC-3000", "dba-team").
			WithMetadata(dbMeta("primary")).
			WithInterfaces(NewInterface("order-sql-primary", "JDBC").
				AtPort(5432).
				WithDB("orders_v1").
				OnHost("orders-primary.example.com")),
	)

	a.AddNode(
		NewNode(
			"order-database-replica",
			Database,
			"Order Database (Replica)",
			"Read-only replica for scaling read operations.",
		).
			SetStandards("CC-3000", "dba-team").
			WithMetadata(dbMeta("replica")).
			WithInterfaces(NewInterface("order-sql-replica", "JDBC").
				AtPort(5432).
				WithDB("orders_v1").
				OnHost("orders-replica.example.com")),
	)

	invDBMeta := NewMetadata().
		Add("backup-schedule", "weekly at Sunday 03:00 UTC").
		Add("dba-contact", "dba-team@example.com").
		Add("deployment-type", "managed-service").
		Add("restore-time", "30 minutes").
		Add("tech-owner", "DBA Team")

	a.AddNode(NewNode("inventory-db", Database, "Inventory Database", "Stores stock levels.").
		SetStandards("CC-4000", "dba-team").
		WithMetadata(invDBMeta).
		WithInterfaces(NewInterface("inventory-sql", "JDBC").
			AtPort(5432).
			WithDB("inventory_v1").
			OnHost("inventory-db.example.com")))
}

func addRelationships(a *Architecture) {
	// Actors
	a.AddRelationship(
		Interacts(
			"customer-interacts-lb",
			"Customer accesses the platform via Load Balancer.",
			"customer",
			"load-balancer",
		).
			WithData("public", true),
	)
	a.AddRelationship(
		Interacts("admin-interacts-lb", "Admin manages the platform via Load Balancer.", "admin", "load-balancer").
			WithData("internal", true),
	)

	// LB to GW
	a.AddRelationship(
		ConnectIntf(
			"lb-connects-gateway-1",
			"Load Balancer distributes traffic to Gateway Instance 1.",
			"load-balancer",
			"api-gateway-1",
			[]string{"lb-to-gateway"},
			[]string{"gateway-1-http"},
		).
			WithData("internal", true),
	)
	a.AddRelationship(
		ConnectIntf(
			"lb-connects-gateway-2",
			"Load Balancer distributes traffic to Gateway Instance 2.",
			"load-balancer",
			"api-gateway-2",
			[]string{"lb-to-gateway"},
			[]string{"gateway-2-http"},
		).
			WithData("internal", true),
	)

	// GW to Svc
	a.AddRelationship(
		ConnectIntf(
			"gateway-1-connects-order",
			"Gateway 1 forwards requests to Order Service.",
			"api-gateway-1",
			"order-service",
			[]string{"order-client-1"},
			[]string{"order-api"},
		).
			WithData("internal", true),
	)
	a.AddRelationship(
		ConnectIntf(
			"gateway-2-connects-order",
			"Gateway 2 forwards requests to Order Service.",
			"api-gateway-2",
			"order-service",
			[]string{"order-client-2"},
			[]string{"order-api"},
		).
			WithData("internal", true),
	)
	a.AddRelationship(
		ConnectIntf(
			"gateway-1-connects-inventory",
			"Gateway 1 forwards requests to Inventory Service.",
			"api-gateway-1",
			"inventory-service",
			[]string{"inventory-client-1"},
			[]string{"inventory-api"},
		).
			WithData("internal", true),
	)
	a.AddRelationship(
		ConnectIntf(
			"gateway-2-connects-inventory",
			"Gateway 2 forwards requests to Inventory Service.",
			"api-gateway-2",
			"inventory-service",
			[]string{"inventory-client-2"},
			[]string{"inventory-api"},
		).
			WithData("internal", true),
	)

	// Internal connections
	a.AddRelationship(
		ConnectIntf(
			"order-connects-primary-db",
			"Order Service persists data to Primary Order Database.",
			"order-service",
			"order-database-primary",
			[]string{"order-db-write-client"},
			[]string{"order-sql-primary"},
		).
			WithData("confidential", true).
			WithMetadata(NewMetadata().Add("monitoring", true)),
	)
	a.AddRelationship(
		ConnectIntf(
			"order-publishes-to-queue",
			"Order Service publishes payment task to queue.",
			"order-service",
			"order-queue",
			[]string{"payment-publisher"},
			nil,
		).
			WithData("internal", true).
			WithProtocol("AMQP"),
	)
	a.AddRelationship(
		ConnectIntf(
			"payment-subscribes-to-queue",
			"Payment Service consumes payment task from queue.",
			"order-queue",
			"payment-service",
			nil,
			[]string{"payment-consumer"},
		).
			WithData("internal", true).
			WithProtocol("AMQP"),
	)
	a.AddRelationship(
		ConnectIntf(
			"order-connects-replica-db",
			"Order Service reads data from Replica Order Database.",
			"order-service",
			"order-database-replica",
			[]string{"order-db-read-client"},
			[]string{"order-sql-replica"},
		).
			WithData("confidential", true).
			WithMetadata(NewMetadata().Add("monitoring", true)),
	)
	a.AddRelationship(
		ConnectIntf(
			"order-connects-inventory",
			"Order Service checks/reserves stock in Inventory Service.",
			"order-service",
			"inventory-service",
			[]string{"order-inventory-client"},
			[]string{"inventory-api"},
		).
			WithData("internal", true).
			WithMetadata(NewMetadata().Add("circuit-breaker", true).Add("latency-sla", "< 100ms").Add("monitoring", true)),
	)
	a.AddRelationship(
		ConnectIntf(
			"inventory-connects-db",
			"Inventory Service manages stock in Inventory Database.",
			"inventory-service",
			"inventory-db",
			[]string{"inventory-db-client"},
			[]string{"inventory-sql"},
		).
			WithData("internal", true).
			WithMetadata(NewMetadata().Add("monitoring", true)),
	)

	// Compositions
	a.AddRelationship(
		ComposedOf(
			"broker-composition",
			"Message broker contains the order queue.",
			"message-broker",
			[]string{"order-queue"},
		).
			WithData("internal", false),
	)
	a.AddRelationship(
		ComposedOf(
			"order-db-composition",
			"Primary and replica databases form the order database cluster.",
			"order-database-cluster",
			[]string{"order-database-primary", "order-database-replica"},
		).
			WithData("confidential", true),
	)
	a.AddRelationship(
		ComposedOf(
			"ecommerce-system-composition",
			"The E-Commerce Platform comprises its core services and databases.",
			"ecommerce-system",
			[]string{
				"load-balancer",
				"api-gateway-1",
				"api-gateway-2",
				"order-service",
				"inventory-service",
				"payment-service",
				"order-database-cluster",
				"inventory-db",
				"message-broker",
			},
		).
			WithData("internal", false),
	)
}
