package main

import "fmt"

func main() {
	arch := NewArchitecture(
		"ecommerce-platform-architecture",
		"E-Commerce Order Processing Platform",
		"A complete architecture for an e-commerce order processing system.",
	)

	setupGeneralInfo(arch)
	addGlobalControls(arch)
	addNodes(arch)
	addRelationships(arch)
	addFlows(arch)

	fmt.Println(arch.ToJSON())
}

func setupGeneralInfo(a *Architecture) {
	a.ADRs = []string{
		"docs/adr/0001-use-message-queue-for-async-processing.md",
		"docs/adr/0002-use-oauth2-for-api-authentication.md",
	}

	monitoring := map[string]any{
		"grafana-dashboard": "https://grafana.example.com/d/ecommerce-overview",
		"kibana-logs":       "https://kibana.example.com/app/discover#/ecommerce-*",
		"pagerduty-service": "https://pagerduty.example.com/services/ECOMMERCE",
		"statuspage":        "https://status.example.com",
		"metrics-retention": "30 days",
		"log-retention":     "90 days",
	}

	a.AddMeta("owner", "Architecture Team").
		AddMeta("version", "1.0.0").
		AddMeta("created", "2025-12-26").
		AddMeta("description", "E-commerce order processing platform architecture demo.").
		AddMeta("tags", []string{"ecommerce", "microservices", "orders"}).
		AddMeta("monitoring", monitoring)
}

func addGlobalControls(a *Architecture) {
	a.AddControl(
		"security",
		"Data encryption and secure communication requirements",
		NewRequirement(
			"https://internal-policy.example.com/security/encryption-at-rest",
			NewSecurityConfig("AES-256", "all-data-stores"),
		),
		NewRequirementURL(
			"https://internal-policy.example.com/security/tls-1-3-minimum",
			"https://configs.example.com/security/tls-config.yaml",
		),
	).
		AddControl(
			"performance",
			"System-wide performance and scalability requirements",
			NewRequirement(
				"https://internal-policy.example.com/performance/response-time-sla",
				NewPerformanceConfig(200, 100),
			),
			NewRequirementURL(
				"https://internal-policy.example.com/performance/availability-target",
				"https://configs.example.com/infra/ha-config.yaml",
			),
		).
		AddControl(
			"high-availability",
			"System-wide uptime and availability requirements",
			NewRequirement(
				"https://internal-policy.example.com/resilience/availability-sla",
				NewAvailabilityConfig(99.9, 60),
			),
		)
}

func addFlows(a *Architecture) {
	flowMeta := map[string]any{
		"business-impact":        "Customers cannot complete purchases - direct revenue loss",
		"degraded-behavior":      "Orders queue in message broker; processed when service recovers",
		"customer-communication": "Display 'Order processing delayed' message",
		"sla":                    "99.9% availability, 30s p99 latency",
	}

	a.Flow("order-processing-flow", "Customer Order Processing", "End-to-end flow from customer placing an order to payment confirmation").
		AddMeta("business-impact", flowMeta["business-impact"]).
		AddMeta("degraded-behavior", flowMeta["degraded-behavior"]).
		AddMeta("customer-communication", flowMeta["customer-communication"]).
		AddMeta("sla", flowMeta["sla"]).
		Step("customer-interacts-lb", 1, "Customer submits order via Load Balancer", "source-to-destination").
		Step("lb-connects-gateway-1", 2, "LB routes to Gateway 1", "source-to-destination").
		Step("gateway-1-connects-order", 3, "Gateway 1 routes to Order Service", "source-to-destination").
		Step("order-publishes-to-queue", 4, "Order Service publishes payment task", "source-to-destination").
		Step("payment-subscribes-to-queue", 5, "Payment Service processes task from queue", "source-to-destination")

	invFlowMeta := map[string]any{
		"business-impact":        "Stock levels may be inaccurate - risk of overselling",
		"degraded-behavior":      "Fall back to cached inventory; flag orders for manual review",
		"customer-communication": "Display 'Stock availability being confirmed'",
		"sla":                    "99.5% availability, 500ms p99 latency",
	}

	a.Flow("inventory-check-flow", "Inventory Stock Check", "Admin checks and updates inventory stock levels").
		AddMeta("business-impact", invFlowMeta["business-impact"]).
		AddMeta("degraded-behavior", invFlowMeta["degraded-behavior"]).
		AddMeta("customer-communication", invFlowMeta["customer-communication"]).
		AddMeta("sla", invFlowMeta["sla"]).
		Step("admin-interacts-lb", 1, "Admin requests inventory status via LB", "source-to-destination").
		Step("lb-connects-gateway-2", 2, "LB routes to Gateway 2", "source-to-destination").
		Step("gateway-2-connects-inventory", 3, "Gateway 2 routes to inventory service", "source-to-destination").
		Step("inventory-connects-db", 4, "Query current stock levels", "source-to-destination").
		Step("inventory-connects-db", 5, "Return stock data", "destination-to-source")
}

func addNodes(a *Architecture) {
	a.Node("customer", Actor, "Customer", "A user who browses and purchases products.").
		Standard("CC-1000", "marketing-team")

	a.Node("admin", Actor, "Admin", "A staff member who manages products and orders.").
		Standard("CC-1000", "ops-team")

	a.Node("ecommerce-system", System, "E-Commerce Platform", "The overall e-commerce system containing microservices.").
		Standard("CC-2000", "platform-team")

	lb := a.Node(
		"load-balancer",
		Service,
		"Load Balancer",
		"High-availability entry point that distributes traffic to API Gateways.",
	)
	lb.Standard("CC-2000", "platform-team").
		AddMeta("tech-owner", "Network Team").
		AddMeta("owner", "platform-team").AddMeta("oncall-slack", "#oncall-platform").
		AddMeta("health-endpoint", "/status").AddMeta("runbook", "https://runbooks.example.com/load-balancer").
		AddMeta("tier", "tier-1").AddMeta("dependencies", []string{"api-gateway-1", "api-gateway-2"}).
		AddMeta("dashboard", "https://grafana.example.com/d/lb-metrics").
		AddMeta("log-query", "service:load-balancer AND error").
		AddMeta("alerts", []string{"LB-HighLatency", "LB-TargetGroupUnhealthy"}).
		AddMeta("deployment-type", "managed-service").
		AddMeta("business-criticality", "high").AddMeta("ha-enabled", true)
	lb.Interface("lb-https", "HTTPS").SetName("Public HTTPS Interface").SetPort(443).SetHost("api.shop.example.com")
	lb.Interface("lb-to-gateway", "HTTP").SetDesc("Outbound to API Gateways")

	gwPerf := []Requirement{
		NewRequirementURL(
			"https://internal-policy.example.com/performance/rate-limiting",
			"https://configs.example.com/gateway/rate-limits.yaml",
		),
		NewRequirement(
			"https://internal-policy.example.com/performance/caching-policy",
			map[string]any{"default-ttl-seconds": 300, "cache-control": "private"},
		),
	}

	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("api-gateway-%d", i)
		name := fmt.Sprintf("API Gateway Instance %d", i)
		desc := "Primary API Gateway instance."
		if i == 2 {
			desc = "Secondary API Gateway instance for high availability."
		}

		gw := a.Node(id, Service, name, desc).
			Standard("CC-2000", "platform-team").
			AddMeta("tech-owner", "Edge Team").
			AddMeta("owner", "platform-team").AddMeta("oncall-slack", "#oncall-platform").
			AddMeta("health-endpoint", "/health").AddMeta("runbook", "https://runbooks.example.com/api-gateway").
			AddMeta("tier", "tier-1").AddMeta("dependencies", []string{"order-service", "inventory-service"}).
			AddMeta("dashboard", "https://grafana.example.com/d/gateway-overview").
			AddMeta("log-query", fmt.Sprintf("service:api-gateway AND instance:%d", i)).
			AddMeta("alerts", []string{"Gateway-5xx-Rate", "Gateway-HighLatency"}).
			AddMeta("repository", "https://github.com/example/api-gateway").
			AddMeta("deployment-type", "container").AddMeta("business-criticality", "high")
		gw.AddControl("performance", "API Gateway rate limiting and caching requirements", gwPerf...)
		gw.Interface(fmt.Sprintf("gateway-%d-http", i), "HTTP").SetName("HTTP Interface").SetPort(80)
		gw.Interface(fmt.Sprintf("order-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("inventory-client-%d", i), "REST")
		gw.Interface(fmt.Sprintf("gateway-%d-health", i), "HTTP").SetName("Health Check").SetPath("/health")
	}

	orderFailures := []map[string]any{
		{
			"check":        "Check connection pool metrics in Grafana dashboard",
			"escalation":   "If persists > 5min, page DBA team",
			"likely-cause": "Database connection pool exhausted",
			"remediation":  "Scale up service replicas or increase pool size",
			"symptom":      "HTTP 503 errors",
		},
		{
			"check":        "Check payment-service health and circuit breaker status",
			"escalation":   "Contact payments-team if circuit breaker not triggering",
			"likely-cause": "Payment service degradation",
			"remediation":  "Circuit breaker should open automatically; check fallback queue",
			"symptom":      "High latency (>2s p99)",
		},
		{
			"check":        "Verify inventory-service cache TTL and database replication lag",
			"escalation":   "Contact platform-team for cache issues",
			"likely-cause": "Inventory service returning stale data",
			"remediation":  "Clear inventory cache; check replica sync status",
			"symptom":      "Order validation failures",
		},
	}

	orderSvc := a.Node("order-service", Service, "Order Service", "Handles order creation and lifecycle management.")
	orderSvc.Standard("CC-3000", "orders-team").
		AddMeta("alerts", []string{"OrderCreationFailureRate", "OrderDBConectionExhausted"}).
		AddMeta("business-criticality", "high").
		AddMeta("dashboard", "https://grafana.example.com/d/order-service-metrics").
		AddMeta("dependencies", []string{"order-database-cluster", "inventory-service", "message-broker"}).
		AddMeta("deployment-type", "container").
		AddMeta("failure-modes", orderFailures).
		AddMeta("health-endpoint", "/actuator/health").
		AddMeta("log-query", "app:order-service AND level:ERROR").
		AddMeta("oncall-slack", "#oncall-orders").
		AddMeta("owner", "orders-team").
		AddMeta("repository", "https://github.com/example/order-service").
		AddMeta("runbook", "https://runbooks.example.com/order-service").
		AddMeta("sla-tier", "tier-1").
		AddMeta("tech-owner", "Order Team").
		AddMeta("tier", "tier-1")
	orderSvc.AddControl(
		"circuit-breaker",
		"Fault tolerance for downstream service calls",
		NewRequirement(
			"https://internal-policy.example.com/resilience/circuit-breaker-policy",
			NewCircuitBreakerConfig(50, 30, 10),
		),
	)
	orderSvc.Interface("order-api", "REST").SetName("Order API").SetPort(8080)
	orderSvc.Interface("order-db-write-client", "JDBC").
		SetName("Order DB Write Client").
		SetDesc("Writes to the primary database.")
	orderSvc.Interface("order-db-read-client", "JDBC").
		SetName("Order DB Read Client").
		SetDesc("Reads from the replica database.")
	orderSvc.Interface("payment-publisher", "AMQP").SetDesc("Publishes order messages for payment processing.")
	orderSvc.Interface("order-inventory-client", "REST").SetDesc("Outbound connection to check inventory")
	orderSvc.Interface("order-health", "HTTP").SetName("Health Check").SetPath("/health")

	invSvc := a.Node("inventory-service", Service, "Inventory Service", "Manages product stock levels.")
	invSvc.Standard("CC-4000", "inventory-team").
		AddMeta("alerts", []string{"InventoryCacheInconsistency", "StockUpdateFailure"}).
		AddMeta("business-criticality", "high").
		AddMeta("dashboard", "https://grafana.example.com/d/inventory-metrics").
		AddMeta("dependencies", []string{"inventory-db"}).
		AddMeta("deployment-type", "container").
		AddMeta("failure-modes", []map[string]any{
			{
				"check":        "Check DB lock metrics and slow query log",
				"escalation":   "Contact DBA team for lock contention",
				"likely-cause": "Deadlock on stock updates",
				"remediation":  "Review transaction isolation level or retry logic",
				"symptom":      "Inventory sync failures",
			},
			{
				"check":        "Verify Redis/Memcached availability and evictions",
				"escalation":   "Contact platform-team for cache infrastructure",
				"likely-cause": "Cache invalidation failure",
				"remediation":  "Flush cache for affected products",
				"symptom":      "Stale stock levels",
			},
		}).
		AddMeta("health-endpoint", "/health").
		AddMeta("log-query", "app:inventory-service").
		AddMeta("oncall-slack", "#oncall-inventory").
		AddMeta("owner", "inventory-team").
		AddMeta("repository", "https://github.com/example/inventory-service").
		AddMeta("runbook", "https://runbooks.example.com/inventory-service").
		AddMeta("tech-owner", "Warehouse Team").
		AddMeta("tier", "tier-2")
	invSvc.Interface("inventory-api", "REST").SetName("Inventory API").SetPort(8081)
	invSvc.Interface("inventory-db-client", "JDBC")
	invSvc.Interface("inventory-health", "HTTP").SetName("Health Check").SetPath("/health")

	paySvc := a.Node("payment-service", Service, "Payment Service", "Integrates with external payment providers.")
	paySvc.Standard("CC-5000", "payments-team").
		AddMeta("alerts", []string{"PaymentGatewayTimeout", "PCIViolationAttempt"}).
		AddMeta("business-criticality", "high").
		AddMeta("dashboard", "https://grafana.example.com/d/payment-metrics").
		AddMeta("dependencies", []string{"external-payment-provider"}).
		AddMeta("deployment-type", "serverless").
		AddMeta("failure-modes", []map[string]any{
			{
				"check":        "Verify external gateway status page",
				"escalation":   "Escalate to provider support",
				"likely-cause": "External payment gateway latency",
				"remediation":  "Enable aggressive retry for idempotent calls",
				"symptom":      "Payment processing timeouts",
			},
			{
				"check":        "Review access logs for unusual patterns",
				"escalation":   "Contact security-team",
				"likely-cause": "API Key leaked or compromised",
				"remediation":  "Rotate API keys immediately",
				"symptom":      "Unauthorized transaction spikes",
			},
		}).
		AddMeta("health-endpoint", "/health").
		AddMeta("log-query", "app:payment-service").
		AddMeta("oncall-slack", "#oncall-payments").
		AddMeta("owner", "payments-team").
		AddMeta("repository", "https://github.com/example/payment-service").
		AddMeta("runbook", "https://runbooks.example.com/payment-service").
		AddMeta("tech-owner", "Payment Team").
		AddMeta("tier", "tier-1")
	paySvc.AddControl(
		"compliance",
		"PCI-DSS compliance for payment processing",
		NewRequirementURL(
			"https://www.pcisecuritystandards.org/documents/PCI-DSS-v4.0",
			"https://configs.example.com/compliance/pci-dss-config.json",
		),
	)
	paySvc.Interface("payment-api", "REST").SetName("Payment Processing API").SetPort(8082)
	paySvc.Interface("payment-consumer", "AMQP").SetDesc("Consumes order messages for payment processing.")
	paySvc.Interface("payment-health", "HTTP").SetName("Health Check").SetPath("/health")

	broker := a.Node(
		"message-broker",
		System,
		"Message Broker (RabbitMQ)",
		"Central messaging system for failure isolation and async processing.",
	)
	broker.Standard("CC-2000", "platform-team").
		AddMeta("adr", "docs/adr/0001-use-message-queue-for-async-processing.md").
		AddMeta("alerts", []string{"RabbitMQQueueSizeHigh", "RabbitMQConsumerDown"}).
		AddMeta("dashboard", "https://grafana.example.com/d/rabbitmq-overview").
		AddMeta("deployment-type", "managed-service").
		AddMeta("health-endpoint", "/health").
		AddMeta("log-query", "service:rabbitmq").
		AddMeta("oncall-slack", "#oncall-platform").
		AddMeta("owner", "platform-team").
		AddMeta("runbook", "https://runbooks.example.com/rabbitmq").
		AddMeta("tech-owner", "Platform Team").
		AddMeta("tier", "tier-1")
	broker.Interface("amqp-port", "AMQP").SetPort(5672)

	a.Node("order-queue", Queue, "Order Payment Queue", "Buffer for orders awaiting payment processing.").
		Standard("CC-3000", "orders-team")

	dbCluster := a.Node(
		"order-database-cluster",
		System,
		"Order Database Cluster",
		"High-availability database cluster for order data.",
	)
	dbCluster.Standard("CC-3000", "dba-team").
		AddControl("failover", "Disaster recovery and failover targets",
			NewRequirement(
				"https://internal-policy.example.com/resilience/disaster-recovery-targets",
				NewFailoverConfig(15, 5, true),
			),
		)

	dbCommonMeta := map[string]any{
		"tech-owner":          "DBA Team",
		"backup-schedule":     "daily at 02:00 UTC",
		"restore-time":        "60 minutes",
		"dba-contact":         "dba-team@example.com",
		"deployment-type":     "managed-service",
		"data-classification": "PII",
		"replication-mode":    "async",
	}

	primary := a.Node(
		"order-database-primary",
		Database,
		"Order Database (Primary)",
		"Main writable database for orders.",
	)
	primary.Standard("CC-3000", "dba-team").
		AddMeta("role", "primary").
		AddMeta("backup-schedule", dbCommonMeta["backup-schedule"]).
		AddMeta("data-classification", dbCommonMeta["data-classification"]).
		AddMeta("dba-contact", dbCommonMeta["dba-contact"]).
		AddMeta("deployment-type", dbCommonMeta["deployment-type"]).
		AddMeta("replication-mode", dbCommonMeta["replication-mode"]).
		AddMeta("restore-time", dbCommonMeta["restore-time"]).
		AddMeta("tech-owner", dbCommonMeta["tech-owner"])
	primary.Interface("order-sql-primary", "JDBC").
		SetPort(5432).
		SetDB("orders_v1").
		SetHost("orders-primary.example.com")

	replica := a.Node(
		"order-database-replica",
		Database,
		"Order Database (Replica)",
		"Read-only replica for scaling read operations.",
	)
	replica.Standard("CC-3000", "dba-team").
		AddMeta("role", "replica").
		AddMeta("backup-schedule", dbCommonMeta["backup-schedule"]).
		AddMeta("data-classification", dbCommonMeta["data-classification"]).
		AddMeta("dba-contact", dbCommonMeta["dba-contact"]).
		AddMeta("deployment-type", dbCommonMeta["deployment-type"]).
		AddMeta("replication-mode", dbCommonMeta["replication-mode"]).
		AddMeta("restore-time", dbCommonMeta["restore-time"]).
		AddMeta("tech-owner", dbCommonMeta["tech-owner"])
	replica.Interface("order-sql-replica", "JDBC").
		SetPort(5432).
		SetDB("orders_v1").
		SetHost("orders-replica.example.com")

	invDB := a.Node("inventory-db", Database, "Inventory Database", "Stores stock levels.")
	invDB.Standard("CC-4000", "dba-team").
		AddMeta("backup-schedule", "weekly at Sunday 03:00 UTC").
		AddMeta("dba-contact", "dba-team@example.com").
		AddMeta("deployment-type", "managed-service").
		AddMeta("restore-time", "30 minutes").
		AddMeta("tech-owner", "DBA Team")
	invDB.Interface("inventory-sql", "JDBC").SetPort(5432).SetDB("inventory_v1").SetHost("inventory-db.example.com")
}

func addRelationships(a *Architecture) {
	a.Interacts("customer-interacts-lb", "Customer accesses the platform via Load Balancer.", "customer", "load-balancer").
		Data("public", true)
	a.Interacts("admin-interacts-lb", "Admin manages the platform via Load Balancer.", "admin", "load-balancer").
		Data("internal", true)

	a.Connect("lb-connects-gateway-1", "Load Balancer distributes traffic to Gateway Instance 1.", "load-balancer", "api-gateway-1").
		SrcIntf("lb-to-gateway").
		DstIntf("gateway-1-http").
		Data("internal", true)
	a.Connect("lb-connects-gateway-2", "Load Balancer distributes traffic to Gateway Instance 2.", "load-balancer", "api-gateway-2").
		SrcIntf("lb-to-gateway").
		DstIntf("gateway-2-http").
		Data("internal", true)

	a.Connect("gateway-1-connects-order", "Gateway 1 forwards requests to Order Service.", "api-gateway-1", "order-service").
		SrcIntf("order-client-1").
		DstIntf("order-api").
		Data("internal", true)
	a.Connect("gateway-2-connects-order", "Gateway 2 forwards requests to Order Service.", "api-gateway-2", "order-service").
		SrcIntf("order-client-2").
		DstIntf("order-api").
		Data("internal", true)
	a.Connect("gateway-1-connects-inventory", "Gateway 1 forwards requests to Inventory Service.", "api-gateway-1", "inventory-service").
		SrcIntf("inventory-client-1").
		DstIntf("inventory-api").
		Data("internal", true)
	a.Connect("gateway-2-connects-inventory", "Gateway 2 forwards requests to Inventory Service.", "api-gateway-2", "inventory-service").
		SrcIntf("inventory-client-2").
		DstIntf("inventory-api").
		Data("internal", true)

	a.Connect("order-connects-primary-db", "Order Service persists data to Primary Order Database.", "order-service", "order-database-primary").
		SrcIntf("order-db-write-client").
		DstIntf("order-sql-primary").
		Data("confidential", true).
		AddMeta("monitoring", true)
	a.Connect("order-publishes-to-queue", "Order Service publishes payment task to queue.", "order-service", "order-queue").
		SrcIntf("payment-publisher").
		WithProtocol("AMQP").
		Data("internal", true)
	a.Connect("payment-subscribes-to-queue", "Payment Service consumes payment task from queue.", "order-queue", "payment-service").
		DstIntf("payment-consumer").
		WithProtocol("AMQP").
		Data("internal", true)
	a.Connect("order-connects-replica-db", "Order Service reads data from Replica Order Database.", "order-service", "order-database-replica").
		SrcIntf("order-db-read-client").
		DstIntf("order-sql-replica").
		Data("confidential", true).
		AddMeta("monitoring", true)
	a.Connect("order-connects-inventory", "Order Service checks/reserves stock in Inventory Service.", "order-service", "inventory-service").
		SrcIntf("order-inventory-client").
		DstIntf("inventory-api").
		Data("internal", true).
		AddMeta("monitoring", true).
		AddMeta("circuit-breaker", true).
		AddMeta("latency-sla", "< 100ms")
	a.Connect("inventory-connects-db", "Inventory Service manages stock in Inventory Database.", "inventory-service", "inventory-db").
		SrcIntf("inventory-db-client").
		DstIntf("inventory-sql").
		Data("internal", true).
		AddMeta("monitoring", true)

	a.ComposedOf("broker-composition", "Message broker contains the order queue.", "message-broker", []string{"order-queue"}).
		Data("internal", false)
	a.ComposedOf("order-db-composition", "Primary and replica databases form the order database cluster.", "order-database-cluster", []string{"order-database-primary", "order-database-replica"}).
		Data("confidential", true)
	a.ComposedOf("ecommerce-system-composition", "The E-Commerce Platform comprises its core services and databases.", "ecommerce-system", []string{"load-balancer", "api-gateway-1", "api-gateway-2", "order-service", "inventory-service", "payment-service", "order-database-cluster", "inventory-db", "message-broker"}).
		Data("internal", false)
}
